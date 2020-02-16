package included

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/vbfox/proto-filter/configuration"
)

type inclusionType int

const (
	inclusionType_unknown inclusionType = iota
	inclusionType_included_implicit
	inclusionType_included_explicit
	inclusionType_excluded_explicit
)

func (s inclusionType) String() string {
	return [...]string{
		"inclusionType_unknown",
		"inclusionType_included_implicit",
		"inclusionType_included_explicit",
		"inclusionType_excluded_explicit",
	}[s]
}

type filterBuilder struct {
	configuration   *configuration.Configuration
	isIncludedCache map[string]configuration.InclusionResult
	inclusionMap    map[string]inclusionType
}

func buildPath(parts []string) string {
	result := ""
	for _, part := range parts {
		if len(result) == 0 {
			result = part
		} else {
			result = result + "/" + part
		}
	}
	return result
}

func (b *filterBuilder) getInclusion(path string) inclusionType {
	existingValue, hasExisting := b.inclusionMap[path]
	if !hasExisting {
		return inclusionType_unknown
	}
	return existingValue
}

func (b *filterBuilder) getIsIncludedFromCache(path string) configuration.InclusionResult {
	existing, ok := b.isIncludedCache[path]
	if ok {
		return existing
	}

	value := b.configuration.IsIncluded(strings.Split(path, "/")...)
	b.isIncludedCache[path] = value
	return value
}

func isIncluded(v inclusionType) bool {
	return v == inclusionType_included_implicit || v == inclusionType_included_explicit
}

type inclusionComputationResult struct {
	configuredInclusion configuration.InclusionResult
	existingValue       inclusionType
	newValue            inclusionType
	needToBeExplored    bool
	childInclude        bool
}

func (b *filterBuilder) computeInclusionType(pathString string, includedByParent bool) (inclusionComputationResult, error) {
	result := inclusionComputationResult{}

	configuredInclusion := b.getIsIncludedFromCache(pathString)
	existingValue := b.getInclusion(pathString)

	result.configuredInclusion = configuredInclusion
	result.existingValue = existingValue
	result.childInclude = includedByParent || (configuredInclusion == configuration.IncludedWithChildren)

	fmt.Printf("\n")
	fmt.Printf("[%v]\n", pathString)
	fmt.Printf("Configured inclusion:%v (Existing = %v)\n", configuredInclusion, existingValue)

	if configuredInclusion == configuration.Excluded {
		if isIncluded(existingValue) {
			return result, fmt.Errorf("Element at path %s was included and is now found as excluded", pathString)
		}

		result.newValue = inclusionType_excluded_explicit
		result.needToBeExplored = false
		return result, nil
	}

	if (configuredInclusion == configuration.IncludedWithChildren) || includedByParent {
		if existingValue == inclusionType_excluded_explicit {
			return result, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		result.newValue = inclusionType_included_explicit
		result.needToBeExplored = !isIncluded(existingValue)
		return result, nil
	}

	if configuredInclusion == configuration.IncludedWithoutChildren {
		if existingValue == inclusionType_excluded_explicit {
			return result, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		if existingValue == inclusionType_included_explicit {
			result.newValue = inclusionType_included_explicit
		} else {
			result.newValue = inclusionType_included_implicit
		}
		result.needToBeExplored = !isIncluded(existingValue)
		return result, nil
	}

	result.newValue = existingValue
	result.needToBeExplored = false
	return result, nil
}

func (b *filterBuilder) includeAny(path []string, includedByParent bool) (bool, bool, error) {
	pathString := buildPath(path)
	result, err := b.computeInclusionType(pathString, includedByParent)
	if err != nil {
		return false, false, err
	}

	fmt.Printf("Inclusion result: %v: %+v\n", result.newValue, result)

	b.inclusionMap[pathString] = result.newValue

	return result.needToBeExplored, result.childInclude, nil
}

func reverseStringSlice(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func getDescriptorPath(descriptor desc.Descriptor) []string {
	result := []string{}
	current := descriptor.GetParent()
	for {
		if current == nil {
			break
		}
		result = append(result, current.GetName())
		current = current.GetParent()
	}
	reverseStringSlice(result)
	return result
}

func (b *filterBuilder) includeField(descriptor *desc.FieldDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	messageType := descriptor.GetMessageType()
	if messageType != nil {
		if messageType.IsMapEntry() {
			keyMessage := descriptor.GetMapKeyType().GetMessageType()
			if keyMessage != nil {
				err = b.includeMessage(keyMessage, getDescriptorPath(keyMessage), childInclude)
			}
			if err == nil {
				valueMessage := descriptor.GetMapValueType().GetMessageType()
				err = b.includeMessage(valueMessage, getDescriptorPath(valueMessage), childInclude)
			}
		} else {
			err = b.includeMessage(messageType, getDescriptorPath(messageType), childInclude)
		}
	}

	if err != nil {
		return fmt.Errorf("Failed to include field %s: %w", currentPath, err)
	}

	return nil
}

func (b *filterBuilder) includeMessage(descriptor *desc.MessageDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())

	if descriptor.IsMapEntry() {
		// Map entry message types are generated for backward compatibility but we ignore them as we handle the map<,>
		// type directly.
		return nil
	}

	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	for _, message := range descriptor.GetNestedMessageTypes() {
		if err := b.includeMessage(message, currentPath, childInclude); err != nil {
			return err
		}
	}

	for _, enum := range descriptor.GetNestedEnumTypes() {
		if err := b.includeEnum(enum, currentPath, childInclude); err != nil {
			return err
		}
	}

	for _, field := range descriptor.GetFields() {
		if err := b.includeField(field, currentPath, childInclude); err != nil {
			return err
		}
	}

	return nil
}

func (b *filterBuilder) includeEnumValue(descriptor *desc.EnumValueDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	_, _, err := b.includeAny(currentPath, includedByParent)
	return err
}

func (b *filterBuilder) includeEnum(descriptor *desc.EnumDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	for _, enumValue := range descriptor.GetValues() {
		if err := b.includeEnumValue(enumValue, currentPath, childInclude); err != nil {
			return err
		}
	}

	return nil
}

func (b *filterBuilder) includeServiceMethod(descriptor *desc.MethodDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	inputType := descriptor.GetInputType()
	err = b.includeMessage(inputType, getDescriptorPath(inputType), childInclude)
	if err != nil {
		return err
	}

	outputType := descriptor.GetOutputType()
	err = b.includeMessage(outputType, getDescriptorPath(outputType), childInclude)
	if err != nil {
		return err
	}

	return nil
}

func (b *filterBuilder) includeService(descriptor *desc.ServiceDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	for _, method := range descriptor.GetMethods() {
		if err := b.includeServiceMethod(method, currentPath, childInclude); err != nil {
			return err
		}
	}

	return nil
}

func (b *filterBuilder) includeFileDescriptor(descriptor *desc.FileDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	ok, childInclude, err := b.includeAny(currentPath, includedByParent)
	if !ok {
		return err
	}

	for _, message := range descriptor.GetMessageTypes() {
		if err := b.includeMessage(message, currentPath, childInclude); err != nil {
			return err
		}
	}

	for _, enum := range descriptor.GetEnumTypes() {
		if err := b.includeEnum(enum, currentPath, childInclude); err != nil {
			return err
		}
	}

	for _, service := range descriptor.GetServices() {
		if err := b.includeService(service, currentPath, childInclude); err != nil {
			return err
		}
	}

	return nil
}

func buildInclusions(descriptors []*desc.FileDescriptor, cfg *configuration.Configuration) (map[string]inclusionType, error) {
	builder := filterBuilder{
		isIncludedCache: make(map[string]configuration.InclusionResult),
		configuration:   cfg,
		inclusionMap:    make(map[string]inclusionType),
	}

	for _, descriptor := range descriptors {
		if err := builder.includeFileDescriptor(descriptor, []string{}, false); err != nil {
			return builder.inclusionMap, err
		}
	}

	return builder.inclusionMap, nil
}

// BuildIncluded create a map of every file, message, enum, field  and service that can be
// encountered and if they are included or not
func BuildIncluded(descriptors []*desc.FileDescriptor, configuration *configuration.Configuration) (map[string]bool, error) {
	result := make(map[string]bool)

	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")

	inclusions, err := buildInclusions(descriptors, configuration)
	if err != nil {
		return result, err
	}

	for path, inclusionType := range inclusions {
		switch inclusionType {
		case inclusionType_included_implicit, inclusionType_included_explicit:
			result[path] = true
		default:
			result[path] = false
		}
	}

	return result, nil
}
