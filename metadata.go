package db

import (
	"github.com/iancoleman/strcase"
	"strings"
	"sync"
)

const (
	String   = "string"
	Int      = "int"
	Float    = "float"
	Bool     = "bool"
	Datetime = "datetime"
	Object   = "object"
	Array    = "array"
)

const (
	FormatPassword      = "password"
	FormatISO           = "iso"
	FormatUnixTimestamp = "unix_timestamp"
)

var (
	metadataMap   = make(map[string]Metadata)
	metadataMapMu sync.RWMutex
)

type Metadata struct {
	source           *Connection
	nativeProperties Fields

	Name        string `valid:"required,!empty"`
	NativeName  string
	DisplayName string
	Description string
	Properties  Fields
}

type MetadataInterface interface {
	Metadata() Metadata
}

func (m *Metadata) Session() *Connection {
	return m.source
}

func (m *Metadata) MustNativeName() string {
	if m.NativeName != "" {
		return m.NativeName
	}
	return strcase.ToSnake(m.Name)
}

func (m *Metadata) FieldByName(name string) (f Field, has bool) {
	f, has = m.Properties[name]
	if !has {
		f, has = m.nativeProperties[name]
	}
	return
}

func (m *Metadata) MustFieldNativeName(name string) string {
	if f, has := m.FieldByName(name); has {
		return f.MustNativeName()
	} else if strings.HasPrefix(name, "!") {
		return name[1:]
	} else {
		return strcase.ToSnake(name)
	}
}

type Fields map[string]Field

func (fields Fields) updateFieldNames() Fields {
	if len(fields) > 0 {
		for k, v := range fields {
			v.Name = k
			if v.NativeName == "" {
				v.NativeName = strcase.ToSnake(v.Name)
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
	Format       string
	Relationship Relationship
}

func (f *Field) MustNativeName() string {
	if f.NativeName != "" {
		return f.NativeName
	}
	return strcase.ToSnake(f.Name)
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

func RegisterMetadata(sourceName string, metadata ...interface{}) error {
	if len(metadata) == 0 {
		return nil
	}

	if sourceName == "" {
		return Errorf("missing data source name")
	}
	if !HasSession(sourceName) {
		return Errorf(`unconnected data source "%s"`, sourceName)
	}
	return Session(sourceName).RegisterMetadata(metadata...)
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
