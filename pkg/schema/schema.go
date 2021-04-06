package schema

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	"gopkg.in/guregu/null.v4"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	schemas   = make(map[string]Schema)
	schemasMu sync.RWMutex
)

type schemaInterface interface {
	Schema() Schema
}

type Schema struct {
	DataSource  string           `db:"dataSource" json:"dataSource"`
	Name        string           `db:"name" json:"name" valid:"required"`
	NativeName  string           `db:"nativeName" json:"nativeName" valid:"required"`
	Title       string           `db:"title" json:"title" valid:"required"`
	Description null.String      `db:"description" json:"description"`
	Properties  map[string]Field `db:"properties" json:"-" valid:"-"`
}

func (s *Schema) FieldByName(name string) (field Field, has bool) {
	field, has = s.Properties[name]
	return
}

func (s *Schema) PrimaryFields() (fields []Field) {
	for _, field := range s.Properties {
		if field.Primary.Bool {
			fields = append(fields, field)
		}
	}
	return
}

func (s *Schema) PrimaryField() (field Field, has bool) {
	fields := s.PrimaryFields()
	if len(fields) > 0 {
		return fields[0], true
	}
	return
}

func (s *Schema) FieldByNativeName(nativeName string) (field Field, has bool) {
	for _, field := range s.Properties {
		if field.NativeName == nativeName {
			return field, true
		}
	}
	return
}

func (s *Schema) clone() *Schema {
	clone := *s
	return &clone
}

func RegisterSchema(s Schema) error {
	schemasMu.Lock()
	defer schemasMu.Unlock()

	if _, err := govalidator.ValidateStruct(&s); err != nil {
		return err
	}

	schemas[s.Name] = s
	return nil
}

func parseStructSchema(i interface{}) (*Schema, error) {
	if i == nil {
		return nil, nil
	}
	indirectValue := reflect.Indirect(reflect.ValueOf(i))
	if indirectValue.Kind() != reflect.Struct {
		return nil, nil
	}

	var (
		properties        = make(map[string]Field)
		indirectValueType = indirectValue.Type()
	)

	for i := 0; i < indirectValueType.NumField(); i++ {
		fieldStruct := indirectValueType.Field(i)

		indirectType := fieldStruct.Type
		for indirectType.Kind() == reflect.Ptr {
			indirectType = indirectType.Elem()
		}
		tag := strings.TrimSpace(fieldStruct.Tag.Get("db"))
		if tag == "-" || tag == "" {
			continue
		}
		settings := parseTag(fieldStruct.Tag)
		fieldValue := reflect.New(indirectType).Interface()
		field := Field{Name: fieldStruct.Name}
		if v, has := settings["TYPE"]; has && v != "" {
			field.Type = strings.ToLower(v)
		} else {
			fieldType, err := parseType(fieldValue, indirectType)
			if err != nil {
				return nil, err
			}
			field.Type = fieldType
		}
		if v, has := settings["NATIVENAME"]; has && v != "" {
			field.NativeName = v
		} else if v, has = settings["NATIVE"]; has && v != "" {
			field.NativeName = v
		}
		if v, has := settings["NATIVETYPE"]; has && v != "" {
			field.NativeType = v
		}
		if v, has := settings["TITLE"]; has && v != "" {
			field.Title = v
		} else {
			field.Title = field.Name
		}
		if v, has := settings["ELEMENTTYPE"]; has && v != "" {
			field.ElementType = null.StringFrom(v)
		}
		if v, has := settings["DESCRIPTION"]; has && v != "" {
			field.Description = null.StringFrom(v)
		}
		if v, has := settings["REQUIRED"]; has {
			if v == "REQUIRED" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Required = null.BoolFrom(b)
		}
		if v, has := settings["DEFAULTVALUE"]; has {
			field.DefaultValue = null.StringFrom(v)
		} else if v, has = settings["DEFAULT"]; has {
			field.DefaultValue = null.StringFrom(v)
		}
		if v, has := settings["FORCEDEFAULTVALUE"]; has {
			if v == "FORCEDEFAULTVALUE" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ForceDefaultValue = null.BoolFrom(b)
		} else if v, has = settings["FORCEDEFAULT"]; has {
			if v == "FORCEDEFAULT" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ForceDefaultValue = null.BoolFrom(b)
		}
		if v, has := settings["PRIMARY"]; has {
			if v == "PRIMARY" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Primary = null.BoolFrom(b)
		}
		if v, has := settings["INDEX"]; has && v != "" {
			field.Index = null.StringFrom(v)
		}
		if v, has := settings["UNIQUE"]; has {
			if v == "UNIQUE" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Unique = null.BoolFrom(b)
		}
		if v, has := settings["AUTOINC"]; has {
			if v == "AUTOINC" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.AutoInc = null.BoolFrom(b)
		}
		if v, has := settings["LOWERCASE"]; has {
			if v == "LOWERCASE" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Lowercase = null.BoolFrom(b)
		}
		if v, has := settings["UPPERCASE"]; has {
			if v == "UPPERCASE" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Uppercase = null.BoolFrom(b)
		}
		if v, has := settings["TRIM"]; has {
			if v == "TRIM" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Trim = null.BoolFrom(b)
		}
		if v, has := settings["FORMAT"]; has && v != "" {
			field.Format = null.StringFrom(v)
		}
		if v, has := settings["PATTERN"]; has && v != "" {
			field.Pattern = null.StringFrom(v)
		}
		if v, has := settings["ENUM"]; has && v != "" {
			// Format 1: enum:1,2,3,5
			// Format 2: enum:male=男,female=女,unknown=未知
			var enum []EnumItem
			values := strings.Split(v, ",")
			for _, item := range values {
				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}
				var label, value string
				if strings.Contains(item, "=") {
					pair := strings.Split(item, "=")
					if len(pair) >= 2 {
						value = strings.TrimSpace(pair[0])
						label = strings.TrimSpace(pair[1])
					}
				} else {
					value = item
					label = value
				}
				enum = append(enum, EnumItem{
					Label: label,
					Value: value,
				})

			}
			field.Enum = enum
		}
		if v, has := settings["MINLENGTH"]; has {
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			field.MinLength = null.IntFrom(int64(i))
		}
		if v, has := settings["MAXLENGTH"]; has {
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			field.MaxLength = null.IntFrom(int64(i))
		}
		if v, has := settings["MIN"]; has && v != "" {
			field.Min = null.StringFrom(v)
		}
		if v, has := settings["MAX"]; has && v != "" {
			field.Max = null.StringFrom(v)
		}
		if v, has := settings["EXCLUSIVEMIN"]; has {
			if v == "EXCLUSIVEMIN" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ExclusiveMin = null.BoolFrom(b)
		} else if v, has = settings["EXMIN"]; has {
			if v == "EXMIN" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ExclusiveMin = null.BoolFrom(b)
		}
		if v, has := settings["EXCLUSIVEMAX"]; has {
			if v == "EXCLUSIVEMAX" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ExclusiveMax = null.BoolFrom(b)
		} else if v, has = settings["EXMAX"]; has {
			if v == "EXMAX" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.ExclusiveMax = null.BoolFrom(b)
		}
		if v, has := settings["REF"]; has && v != "" {
			field.Ref = null.StringFrom(v)
		}
		if v, has := settings["OWNER"]; has {
			if v == "OWNER" {
				v = "true"
			}
			b, err := strconv.ParseBool(v)
			if err != nil {
				return nil, err
			}
			field.Owner = null.BoolFrom(b)
		}
		if v, has := settings["LOCALKEY"]; has && v != "" {
			field.LocalKey = null.StringFrom(v)
		}
		if v, has := settings["FOREIGNKEY"]; has && v != "" {
			field.ForeignKey = null.StringFrom(v)
		}
		if v, has := settings["LABEL"]; has && v != "" {
			field.Label = null.StringFrom(v)
		}
		if v, has := settings["ORDER"]; has && v != "" {
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			field.Order = null.IntFrom(int64(i))
		}
		if v, has := settings["GROUP"]; has && v != "" {
			field.Group = null.StringFrom(v)
		}
		properties[field.Name] = field
	}
	schema := Schema{
		Name:       indirectValueType.Name(),
		Properties: properties,
	}
	// 合并定义
	if v, ok := i.(schemaInterface); ok {
		if err := mergo.Merge(&schema, v.Schema()); err != nil {
			return nil, err
		}
	}
	if schema.NativeName == "" {
		schema.NativeName = StructTableName(indirectValueType, false)
	}
	if schema.Title == "" {
		schema.Title = schema.Name
	}
	return &schema, nil
}

func RegisterStructs(s ...interface{}) error {
	for _, i := range s {
		schema, err := parseStructSchema(i)
		if err != nil {
			return err
		}
		err = RegisterSchema(*schema)
		if err != nil {
			return err
		}
	}
	return nil
}

func Metadata(name string) (s Schema, has bool) {
	schemasMu.RLock()
	defer schemasMu.RUnlock()

	s, has = schemas[name]
	return
}

func parseType(i interface{}, indirectType reflect.Type) (string, error) {
	switch i.(type) {
	case int, *int, int8, *int8, int16, *int16, int32, *int32, int64, *int64,
		uint, *uint, uint8, *uint8, uint16, *uint16, uint32, *uint32, uint64, *uint64, uintptr, *uintptr, null.Int, *null.Int:
		return TypeInt, nil
	case float32, *float32, float64, *float64, complex64, *complex64, complex128, *complex128, null.Float, *null.Float:
		return TypeFloat, nil
	case bool, *bool, null.Bool, *null.Bool:
		return TypeBool, nil
	case string, *string, null.String, *null.String:
		return TypeString, nil
	case time.Time, *time.Time, null.Time, *null.Time:
		return TypeTime, nil
	default:
		kindString := indirectType.Kind().String()
		switch kindString {
		case "struct", "map", "chan", "func", "ptr", "unsafe.Pointer", "interface":
			return TypeObject, nil
		case "slice", "array":
			return TypeArray, nil
		default:
			return "", fmt.Errorf("unsupported type: %s", kindString)
		}
	}
}

func parseTag(tag reflect.StructTag) map[string]string {
	settings := make(map[string]string)
	for _, str := range []string{tag.Get("db")} {
		if str == "" {
			continue
		}
		tags := strings.Split(str, ";")
		for _, value := range tags {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if len(v) >= 2 {
				settings[k] = strings.Join(v[1:], ":")
			} else {
				settings[k] = k
			}
		}
	}
	return settings
}
