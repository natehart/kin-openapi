package jsoninfo

import (
	"reflect"
	"sort"
	"sync"
)

var (
	typeInfos      = map[reflect.Type]*TypeInfo{}
	typeInfosMutex sync.RWMutex
)

// TypeInfo contains information about JSON serialization of a type
type TypeInfo struct {
	mutex          sync.RWMutex
	MultipleFields bool // Whether multiple Go fields share the same JSON name
	Type           reflect.Type
	Fields         []FieldInfo
	extensions     []*ExtensionInfo
	Schema         interface{}
	SchemaMutex    sync.RWMutex
}

func GetTypeInfoForValue(value interface{}) *TypeInfo {
	return GetTypeInfo(reflect.TypeOf(value))
}

// GetTypeInfo returns TypeInfo for the given type.
func GetTypeInfo(t reflect.Type) *TypeInfo {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	typeInfosMutex.RLock()
	typeInfo, exists := typeInfos[t]
	typeInfosMutex.RUnlock()
	if exists {
		return typeInfo
	}
	if t.Kind() != reflect.Struct {
		typeInfo = &TypeInfo{
			Type: t,
		}
	} else {
		// Allocate
		typeInfo = &TypeInfo{
			Type:   t,
			Fields: make([]FieldInfo, 0, 16),
		}

		// Add fields
		typeInfo.Fields = AppendFields(nil, nil, t)

		// Sort fields
		sort.Sort(sortableFieldInfos(typeInfo.Fields))
	}

	// Publish
	typeInfosMutex.Lock()
	typeInfos[t] = typeInfo
	typeInfosMutex.Unlock()
	return typeInfo
}

func (typeInfo *TypeInfo) AddExtension(ext *ExtensionInfo) {
	typeInfo.mutex.Lock()
	defer typeInfo.mutex.Unlock()
	typeInfo.extensions = append(typeInfo.extensions, ext)
}

func (typeInfo *TypeInfo) Extensions() []*ExtensionInfo {
	typeInfo.mutex.RLock()
	defer typeInfo.mutex.RUnlock()
	return typeInfo.extensions
}

// FieldNames returns all field names
func (typeInfo *TypeInfo) FieldNames() []string {
	fields := typeInfo.Fields
	names := make([]string, len(fields))
	for i, field := range fields {
		names[i] = field.JSONName
	}
	return names
}
