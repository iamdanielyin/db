package db

import (
	"bytes"
	"github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	"github.com/yuyitech/db/internal/safe_map"
	"github.com/yuyitech/db/pkg/logger"
	"gopkg.in/guregu/null.v4"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	metadata        = make(map[string]Metadata)
	metadataMu      sync.RWMutex
	metaNamingCache = safe_map.NewSafeMapString()
)

type IMetadata interface {
	Metadata() Metadata
}

func ConvertMetadataName(name string) string {
	if v := metaNamingCache.Get(name); v != "" {
		return v
	}

	if name == "" {
		return ""
	}

	var (
		b     bytes.Buffer
		width int
	)

	for i, r := range name {
		if r == '_' || r == '-' || r == '=' || r == ' ' ||
			(r <= 0x7F && ('0' <= r && r <= '9')) {
			width++
			continue
		}
		if width > 0 || i == 0 || ('A' <= r && r <= 'Z') {
			b.WriteRune(unicode.ToUpper(r))
			width = 0
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}

	s := b.String()
	metaNamingCache.Set(name, s)

	return s
}

func parseFieldType(fv interface{}, indirectType reflect.Type) string {
	switch fv.(type) {
	case int, *int, int8, *int8, int16, *int16, int32, *int32, int64, *int64,
		uint, *uint, uint8, *uint8, uint16, *uint16, uint32, *uint32, uint64, *uint64, uintptr, *uintptr, null.Int, *null.Int:
		return TypeInt
	case float32, *float32, float64, *float64, complex64, *complex64, complex128, *complex128, null.Float, *null.Float:
		return TypeFloat
	case bool, *bool, null.Bool, *null.Bool:
		return TypeBool
	case string, *string, null.String, *null.String:
		return TypeString
	case time.Time, *time.Time, null.Time, *null.Time:
		return TypeTime
	default:
		kindString := indirectType.Kind().String()
		switch kindString {
		case "struct", "map", "chan", "func", "ptr", "unsafe.Pointer", "interface":
			return TypeObject
		case "slice", "array":
			return TypeArray
		}
	}
	return ""
}

func ParseMetadata(model interface{}) (*Metadata, error) {
	if model == nil {
		return nil, nil
	}
	if v, ok := model.(Metadata); ok {
		return &v, nil
	} else if v, ok := model.(*Metadata); ok {
		return v, nil
	}
	reflectValue := reflect.Indirect(reflect.ValueOf(model))
	if reflectValue.Kind() != reflect.Struct {
		return nil, nil
	}

	var (
		fields       = make(map[string]Field)
		reflectType  = reflectValue.Type()
		stringFields = make(map[string]string)
	)
	for i := 0; i < reflectType.NumField(); i++ {
		fieldStruct := reflectType.Field(i)

		indirectType := fieldStruct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}

		jsonName := strings.ReplaceAll(fieldStruct.Tag.Get("json"), ",omitempty", "")
		if jsonName == "-" {
			continue
		}
		if jsonName == "" {
			jsonName = fieldStruct.Name
		}
		validRequired := strings.Contains(fieldStruct.Tag.Get("valid"), "required")
		field := Field{
			Name:       jsonName,
			NativeName: jsonName,
			IsRequired: validRequired,
		}
		fieldValue := reflect.New(indirectType).Interface()
		field.Type = parseFieldType(fieldValue, indirectType)
		switch field.Type {
		case TypeObject:
			sub, err := ParseMetadata(fieldValue)
			field.ElemType = parseFieldType(fieldValue, reflect.Indirect(reflect.ValueOf(fieldValue)).Type())
			if err == nil && sub != nil && sub.Fields != nil && len(sub.Fields) > 0 {
				sub.RegisterChildrenFields(field.Name, sub.Fields)
			}
		case TypeArray:
			elemValue := reflect.MakeSlice(indirectType, 1, 1).Index(0).Interface()
			field.ElemType = parseFieldType(elemValue, reflect.Indirect(reflect.ValueOf(elemValue)).Type())
			if field.ElemType == TypeObject {
				sub, err := ParseMetadata(elemValue)
				if err == nil && sub != nil && sub.Fields != nil && len(sub.Fields) > 0 {
					sub.RegisterChildrenFields(field.Name, sub.Fields)
				}
			}
		}
		fields[field.Name] = field
		if v := fieldStruct.Tag.Get("db"); v != "" {
			stringFields[field.Name] = v
		}
	}
	if len(stringFields) > 0 {
		fieldConfigs := new(Metadata).FieldsFromString(stringFields)
		for k, v := range fieldConfigs {
			raw, has := fields[k]
			if has {
				if err := mergo.Merge(&raw, v); err != nil {
					return nil, err
				}
				fields[k] = raw
			} else {
				fields[k] = v
			}
		}
	}
	meta := Metadata{
		Name:   reflectType.Name(),
		Fields: fields,
	}
	if v, ok := model.(IMetadata); ok {
		src := v.Metadata()
		if src.Fields != nil && len(src.Fields) > 0 {
			for k, v := range src.Fields {
				f, has := meta.Fields[k]
				if has {
					err := mergo.Merge(f, v)
					if err != nil {
						return nil, err
					}
				} else {
					meta.Fields[k] = v
				}
			}
		}
		if err := mergo.Merge(&meta, &src); err != nil {
			return nil, err
		}
	}
	return &meta, nil
}

func RegisterModel(v interface{}) error {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	var meta Metadata
	if a, ok := v.(Metadata); ok {
		meta = a
	} else if a, ok := v.(*Metadata); ok {
		meta = *a
	} else {
		v, err := ParseMetadata(v)
		if err != nil {
			return err
		}
		meta = *v
	}
	if meta.Name == "" {
		return nil
	}

	if _, err := govalidator.ValidateStruct(&meta); err != nil {
		return err
	}
	if meta.NativeName == "" {
		meta.NativeName = meta.Name
	}
	if meta.DataSourceName == "" {
		meta.DataSourceName = defaultDataSourceName
	}

	metadata[meta.Name] = meta
	return nil
}

func UnregisterModel(name string) {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	if _, has := metadata[name]; has {
		delete(metadata, name)
	}
}

func Meta(name string) (Metadata, bool) {
	metadataMu.RLock()
	defer metadataMu.RUnlock()

	meta, has := metadata[name]
	return meta, has
}

func Model(name string) IModel {
	meta, has := Meta(name)
	if !has || meta.Name == "" {
		return nil
	}

	d := DB(meta.DataSourceName)
	if d == nil {
		dsn := meta.DataSourceName
		if dsn == "" {
			dsn = "default"
		}
		logger.ERROR("Data source '%s' is not registered", dsn)
		return nil
	}
	m := d.Model(name)
	if m == nil {
		logger.ERROR("Model '%s' is not registered", name)
	}
	return m
}
