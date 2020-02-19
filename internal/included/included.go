// Package included is an intermediate step in the generation of filtered protobuf.
//
// The target is to take a configuration (Loaded from a configuration file) and a set of
// protobuf file descriptors and generate a set of all the things (files, messages, enums, fields, ...)
// that should be included in the filtered output.
package included

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/desc"
	"github.com/vbfox/proto-filter/configuration"
	"github.com/vbfox/proto-filter/internal/utils"
)

type inclusionType int

const (
	inclusionTypeUnknown inclusionType = iota
	inclusionTypeIncludedImplicit
	inclusionTypeIncludedExplicit
	inclusionTypeExcludedExplicit
)

func (s inclusionType) String() string {
	return [...]string{
		"inclusionType_unknown",
		"inclusionType_included_implicit",
		"inclusionType_included_explicit",
		"inclusionType_excluded_explicit",
	}[s]
}

type IncludedFileDescriptor struct {
	path       string
	descriptor *desc.FileDescriptor
	messages   []*IncludedMessageDescriptor
	enums      []*IncludedEnumDescriptor
	services   []*IncludedServiceDescriptor
}

func (d IncludedFileDescriptor) GetPath() string                { return d.path }
func (d IncludedFileDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedMessageDescriptor struct {
	path       string
	descriptor *desc.MessageDescriptor
	fields     []*IncludedFieldDescriptor
	nested     []*IncludedMessageDescriptor
	enums      []*IncludedEnumDescriptor
}

func (d IncludedMessageDescriptor) GetPath() string                { return d.path }
func (d IncludedMessageDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedFieldDescriptor struct {
	path       string
	descriptor *desc.FieldDescriptor
}

func (d IncludedFieldDescriptor) GetPath() string                { return d.path }
func (d IncludedFieldDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedEnumDescriptor struct {
	path       string
	descriptor *desc.EnumDescriptor
	values     []*IncludedEnumValueDescriptor
}

func (d IncludedEnumDescriptor) GetPath() string                { return d.path }
func (d IncludedEnumDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedEnumValueDescriptor struct {
	path       string
	descriptor *desc.EnumValueDescriptor
}

func (d IncludedEnumValueDescriptor) GetPath() string                { return d.path }
func (d IncludedEnumValueDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedServiceDescriptor struct {
	path       string
	descriptor *desc.ServiceDescriptor
	methods    []*IncludedMethodDescriptor
}

func (d IncludedServiceDescriptor) GetPath() string                { return d.path }
func (d IncludedServiceDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedMethodDescriptor struct {
	path       string
	descriptor *desc.MethodDescriptor
}

func (d IncludedMethodDescriptor) GetPath() string                { return d.path }
func (d IncludedMethodDescriptor) GetDescriptor() desc.Descriptor { return d.descriptor }

type IncludedDescriptor interface {
	GetPath() string
	GetDescriptor() desc.Descriptor
}

type descriptorInclusionInfo struct {
	inclusion  inclusionType
	publicInfo IncludedDescriptor
}

type filterBuilder struct {
	configuration   *configuration.Configuration
	isIncludedCache map[string]configuration.InclusionResult
	inclusionMap    map[string]inclusionType
}

func (b *filterBuilder) getInclusion(path string) inclusionType {
	existingValue, hasExisting := b.inclusionMap[path]
	if !hasExisting {
		return inclusionTypeUnknown
	}
	return existingValue
}

// getIsIncludedFromCache returns how the configuration see the file
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
	return v == inclusionTypeIncludedImplicit || v == inclusionTypeIncludedExplicit
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

		result.newValue = inclusionTypeExcludedExplicit
		result.needToBeExplored = false
		return result, nil
	}

	if (configuredInclusion == configuration.IncludedWithChildren) || includedByParent {
		if existingValue == inclusionTypeExcludedExplicit {
			return result, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		result.newValue = inclusionTypeIncludedExplicit
		result.needToBeExplored = !isIncluded(existingValue)
		return result, nil
	}

	if configuredInclusion == configuration.IncludedWithoutChildren {
		if existingValue == inclusionTypeExcludedExplicit {
			return result, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		if existingValue == inclusionTypeIncludedExplicit {
			result.newValue = inclusionTypeIncludedExplicit
		} else {
			result.newValue = inclusionTypeIncludedImplicit
		}
		result.needToBeExplored = !isIncluded(existingValue)
		return result, nil
	}

	result.newValue = existingValue
	result.needToBeExplored = false
	return result, nil
}

func (b *filterBuilder) includeAny(descriptor desc.Descriptor, pathString string, includedByParent bool) (bool, bool, error) {
	result, err := b.computeInclusionType(pathString, includedByParent)
	if err != nil {
		return false, false, err
	}

	fmt.Printf("Inclusion result: %v: %+v\n", result.newValue, result)

	b.inclusionMap[pathString] = result.newValue

	return result.needToBeExplored, result.childInclude, nil
}

func (b *filterBuilder) includeField(descriptor *desc.FieldDescriptor, path []string, includedByParent bool) (*IncludedFieldDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)

	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedFieldDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	messageType := descriptor.GetMessageType()
	if messageType != nil {
		if messageType.IsMapEntry() {
			keyMessage := descriptor.GetMapKeyType().GetMessageType()
			if keyMessage != nil {
				err = b.includeMessageNonRooted(keyMessage, childInclude)
			}
			if err == nil {
				valueMessage := descriptor.GetMapValueType().GetMessageType()
				err = b.includeMessageNonRooted(valueMessage, childInclude)
			}
		} else {
			err = b.includeMessageNonRooted(messageType, childInclude)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to include field %s: %w", currentPath, err)
	}

	return includedDescriptor, nil
}

func (b *filterBuilder) includeMessage(descriptor *desc.MessageDescriptor, path []string, includedByParent bool) (*IncludedMessageDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)

	if descriptor.IsMapEntry() {
		// Map entry message types are generated for backward compatibility but we ignore them as we handle the map<,>
		// type directly.
		return nil, nil
	}

	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedMessageDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	for _, message := range descriptor.GetNestedMessageTypes() {
		d, err := b.includeMessage(message, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.nested = append(includedDescriptor.nested, d)
		}
	}

	for _, enum := range descriptor.GetNestedEnumTypes() {
		d, err := b.includeEnum(enum, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.enums = append(includedDescriptor.enums, d)
		}
	}

	for _, field := range descriptor.GetFields() {
		d, err := b.includeField(field, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.fields = append(includedDescriptor.fields, d)
		}
	}

	return includedDescriptor, nil
}

// includeMessageNonRooted allow to include a message even if the code has no idea if the parents exists or not
func (b *filterBuilder) includeMessageNonRooted(descriptor *desc.MessageDescriptor, includedByParent bool) error {
	//utils.GetDescriptorPath(descriptor)
	return nil
}

func (b *filterBuilder) includeEnumValue(descriptor *desc.EnumValueDescriptor, path []string, includedByParent bool) (*IncludedEnumValueDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)
	ok, _, err := b.includeAny(descriptor, pathString, includedByParent)

	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedEnumValueDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	return includedDescriptor, nil
}

func (b *filterBuilder) includeEnum(descriptor *desc.EnumDescriptor, path []string, includedByParent bool) (*IncludedEnumDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(path)
	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedEnumDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	for _, enumValue := range descriptor.GetValues() {
		d, err := b.includeEnumValue(enumValue, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.values = append(includedDescriptor.values, d)
		}
	}

	return includedDescriptor, nil
}

func (b *filterBuilder) includeServiceMethod(descriptor *desc.MethodDescriptor, path []string, includedByParent bool) (*IncludedMethodDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)

	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedMethodDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	inputType := descriptor.GetInputType()
	err = b.includeMessageNonRooted(inputType, childInclude)
	if err != nil {
		return nil, err
	}

	outputType := descriptor.GetOutputType()
	err = b.includeMessageNonRooted(outputType, childInclude)
	if err != nil {
		return nil, err
	}

	return includedDescriptor, nil
}

func (b *filterBuilder) includeService(descriptor *desc.ServiceDescriptor, path []string, includedByParent bool) (*IncludedServiceDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)

	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedServiceDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	for _, method := range descriptor.GetMethods() {
		d, err := b.includeServiceMethod(method, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.methods = append(includedDescriptor.methods, d)
		}
	}

	return includedDescriptor, nil
}

func (b *filterBuilder) includeFileDescriptor(descriptor *desc.FileDescriptor, path []string, includedByParent bool) (*IncludedFileDescriptor, error) {
	currentPath := append(path, descriptor.GetName())
	pathString := utils.BuildPath(currentPath)

	ok, childInclude, err := b.includeAny(descriptor, pathString, includedByParent)
	if !ok {
		return nil, err
	}

	includedDescriptor := &IncludedFileDescriptor{
		path:       pathString,
		descriptor: descriptor,
	}

	for _, message := range descriptor.GetMessageTypes() {
		d, err := b.includeMessage(message, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.messages = append(includedDescriptor.messages, d)
		}
	}

	for _, enum := range descriptor.GetEnumTypes() {
		d, err := b.includeEnum(enum, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.enums = append(includedDescriptor.enums, d)
		}
	}

	for _, service := range descriptor.GetServices() {
		d, err := b.includeService(service, currentPath, childInclude)
		if err != nil {
			return nil, err
		}
		if d != nil {
			includedDescriptor.services = append(includedDescriptor.services, d)
		}
	}

	return includedDescriptor, nil
}

func buildInclusions(descriptors []*desc.FileDescriptor, cfg *configuration.Configuration) (map[string]inclusionType, []*IncludedFileDescriptor, error) {
	builder := filterBuilder{
		isIncludedCache: make(map[string]configuration.InclusionResult),
		configuration:   cfg,
		inclusionMap:    make(map[string]inclusionType),
	}

	includedDescriptors := []*IncludedFileDescriptor{}
	for _, descriptor := range descriptors {
		d, err := builder.includeFileDescriptor(descriptor, []string{}, false)
		if err != nil {
			return builder.inclusionMap, nil, err
		}
		if d != nil {
			includedDescriptors = append(includedDescriptors, d)
		}
	}

	return builder.inclusionMap, includedDescriptors, nil
}

// BuildIncluded create a map of every file, message, enum, field  and service that can be
// encountered and if they are included or not
func BuildIncluded(descriptors []*desc.FileDescriptor, configuration *configuration.Configuration) (map[string]inclusionType, error) {
	result := make(map[string]inclusionType)

	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")
	fmt.Printf("==================================================================\n")

	inclusions, _, err := buildInclusions(descriptors, configuration)
	if err != nil {
		return result, err
	}

	for path, inclusionInfo := range inclusions {
		switch inclusionInfo {
		case inclusionTypeIncludedImplicit, inclusionTypeIncludedExplicit:
			result[path] = inclusionInfo
		}
	}

	return result, nil
}
