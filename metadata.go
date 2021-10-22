package db

import (
	"strings"
	"sync"
)

const (
	String    = "string"
	Password  = "password"
	Int       = "int"
	Float     = "float"
	Bool      = "bool"
	Timestamp = "timestamp"
	Time      = "time"
	Object    = "object"
	Array     = "array"
	//ArrayString    = "[string]"
	//ArrayInt       = "[int]"
	//ArrayFloat     = "[float]"
	//ArrayBool      = "[bool]"
	//ArrayTimestamp = "[timestamp]"
	//ArrayObject    = "[object]"
	//ArrayArray     = "[array]"
)

const (
	HookBeforeSave   = "before:save"
	HookBeforeCreate = "before:create"
	HookAfterCreate  = "after:create"
	HookAfterSave    = "after:save"
	HookBeforeUpdate = "before:update"
	HookAfterUpdate  = "after:update"
	HookBeforeFind   = "before:find"
	HookAfterFind    = "after:find"
	HookBeforeDelete = "before:delete"
	HookAfterDelete  = "after:delete"
)

var (
	metadataMap   = make(map[string]Metadata)
	metadataMapMu sync.RWMutex
)

type Metadata struct {
	source           *Connection
	nativeProperties Fields
	hooks            []*MetadataHook
	logicDeleteRule  *LogicDeleteRule

	Name        string `valid:"required,!empty"`
	NativeName  string
	DisplayName string
	Description string
	Properties  Fields
}

type MetadataInterface interface {
	Metadata() Metadata
}

func (m Metadata) Session() *Connection {
	return m.source
}

func (m Metadata) MustNativeName() string {
	if m.NativeName != "" {
		return m.NativeName
	}
	return m.Name
}

func (m Metadata) FieldByName(name string) (f Field, has bool) {
	f, has = m.Properties[name]
	if !has {
		f, has = m.nativeProperties[name]
	}
	return
}

type Fields map[string]Field

func (fields Fields) updateFieldNames() Fields {
	if len(fields) > 0 {
		for k, v := range fields {
			v.Name = k
			if v.NativeName == "" {
				v.NativeName = strings.ToLower(v.Name)
			}
			if len(v.Properties) > 0 {
				v.Properties = v.Properties.updateFieldNames()
			}
			fields[k] = v
		}
	}
	return fields
}

func (fields Fields) nativeFields() Fields {
	nativeProps := make(map[string]Field)
	if len(fields) > 0 {
		for _, v := range fields {
			if len(v.Properties) > 0 {
				v.nativeProperties = v.Properties.nativeFields()
			}
			nativeName := v.MustNativeName()
			nativeProps[nativeName] = v
		}
	}
	return nativeProps
}

type Field struct {
	nativeProperties Fields

	Type         string `valid:"required,!empty"`
	Name         string
	NativeName   string
	DisplayName  string
	Description  string
	Enum         Enum
	Properties   Fields
	Trim         string
	Primary      string
	Required     string
	Unique       string
	DefaultValue string
}

func (f *Field) MustNativeName() string {
	if f.NativeName != "" {
		return f.NativeName
	}
	return f.Name
}

type Enum []EnumItem

func (e Enum) ItemByValue(value string) (v EnumItem, has bool) {
	for _, item := range e {
		if item.Value == value {
			v = item
			has = true
		}
	}
	return
}

type EnumItem struct {
	Label string
	Value string
}

func RegisterMetadata(sourceName string, metadata interface{}) error {
	if metadata == nil {
		return nil
	}

	if sourceName == "" {
		return Errorf("missing data source name")
	}
	if !HasSession(sourceName) {
		return Errorf(`unconnected data source "%s"`, sourceName)
	}
	return Session(sourceName).RegisterMetadata(metadata)
}

func UnregisterMetadata(name string) {
	metadataMapMu.Lock()
	defer metadataMapMu.Unlock()

	name = strings.TrimSpace(name)
	delete(metadataMap, name)
}

func LookupMetadata(name string) (meta Metadata, err error) {
	metadataMapMu.RLock()
	defer metadataMapMu.RUnlock()

	name = strings.TrimSpace(name)
	if v, has := metadataMap[name]; !has {
		err = Errorf(`unregistered metadata "%s"`, name)
		return
	} else {
		meta = v
	}

	return
}
