package included

import (
	"fmt"

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

type filterBuilder struct {
	configuration *configuration.Configuration
	inclusionMap  map[string]inclusionType
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

func isIncluded(v inclusionType) bool {
	return v == inclusionType_included_implicit || v == inclusionType_included_explicit
}

func (b *filterBuilder) includeAny(path []string, includedByParent bool) (bool, bool, error) {
	pathString := buildPath(path)
	configuredInclusion := b.configuration.IsIncluded(path...)

	existingValue := b.getInclusion(pathString)

	if configuredInclusion == configuration.Excluded {
		if isIncluded(existingValue) {
			return false, false, fmt.Errorf("Element at path %s was included and is now found as excluded", pathString)
		}

		b.inclusionMap[pathString] = inclusionType_excluded_explicit
		return false, false, nil
	}

	if (configuredInclusion == configuration.IncludedWithChildren) || includedByParent {
		if existingValue == inclusionType_excluded_explicit {
			return false, false, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		b.inclusionMap[pathString] = inclusionType_included_explicit
	}

	if configuredInclusion == configuration.IncludedWithoutChildren {
		if existingValue == inclusionType_excluded_explicit {
			return false, false, fmt.Errorf("Element at path %s was excluded and is now included", pathString)
		}

		b.inclusionMap[pathString] = inclusionType_included_implicit
	}

	childInclude := includedByParent || (configuredInclusion == configuration.IncludedWithChildren)
	return true, childInclude, nil
}

func (b *filterBuilder) includeField(descriptor *desc.FieldDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
	_, _, err := b.includeAny(currentPath, includedByParent)
	return err
}

func (b *filterBuilder) includeMessage(descriptor *desc.MessageDescriptor, path []string, includedByParent bool) error {
	currentPath := append(path, descriptor.GetName())
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

	return nil
}

func buildInclusions(descriptors []*desc.FileDescriptor, configuration *configuration.Configuration) (map[string]inclusionType, error) {
	builder := filterBuilder{
		configuration: configuration,
		inclusionMap:  make(map[string]inclusionType),
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
