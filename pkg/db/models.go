package db

import (
	"bytes"
	"fmt"
	"github.com/yuyitech/db/pkg/cache"
	"strings"
)

type DataSource struct {
	db    IDatabase
	cache cache.ICache

	Name      string `json:"name"`
	Adapter   string `json:"adapter"`
	DSN       string `json:"dsn"`
	IsDefault bool   `json:"is_default"`
}

type Metadata struct {
	DataSourceName string           `json:"data_source_name"`
	Name           string           `json:"name"`
	NativeName     string           `json:"native_name"`
	DisplayName    string           `json:"display_name"`
	Fields         map[string]Field `json:"fields" valid:"-"`
}

func (m *Metadata) NativeFields() map[string]Field {
	nativeFields := make(map[string]Field)
	for _, v := range m.Fields {
		nativeFields[v.MustNativeName()] = v
	}
	return nativeFields
}

func (m *Metadata) PrimaryFields() (fields []Field) {
	for _, v := range m.Fields {
		if v.IsPrimaryKey {
			fields = append(fields, v)
		}
	}
	return
}

func (m *Metadata) RegisterFields(fields map[string]Field) {
	m.RegisterChildrenFields("", fields)
}

func (m *Metadata) RegisterChildrenFields(parentName string, fields map[string]Field) {
	for k, v := range fields {
		v.MetadataName = m.Name
		if v.NativeName == "" {
			v.NativeName = v.Name
		}
		if parentName != "" && !strings.HasPrefix(fmt.Sprintf("%s.", parentName), v.Name) {
			v.Name = fmt.Sprintf("%s.%s", parentName, v.Name)
		}
		m.Fields[k] = v
	}
}

func (m *Metadata) childrenFields(parentName string, removeParentPrefix bool, level []int, allChildren bool) map[string]Field {
	var l int
	if len(level) > 0 {
		l = level[0]
	} else {
		l = 1
	}
	children := make(map[string]Field)
	parentPrefix := fmt.Sprintf("%s.", parentName)
	for k, v := range m.Fields {
		if !strings.HasPrefix(k, parentPrefix) {
			continue
		}
		kk := strings.Replace(k, parentPrefix, "", 1)
		if l <= 0 {
			if removeParentPrefix {
				children[kk] = v
			} else {
				children[k] = v
			}
		} else {
			split := strings.Split(kk, ".")
			if allChildren {
				if len(split) >= l {
					if removeParentPrefix {
						children[kk] = v
					} else {
						children[k] = v
					}
				}
			} else {
				if len(split) == l {
					if removeParentPrefix {
						children[kk] = v
					} else {
						children[k] = v
					}
				}
			}
		}
	}
	return children
}

func (m *Metadata) ChildrenFields(parentName string, removeParentPrefix bool, level ...int) map[string]Field {
	return m.childrenFields(parentName, removeParentPrefix, level, false)
}

func (m *Metadata) AllChildrenFields(parentName string, removeParentPrefix bool, level ...int) map[string]Field {
	return m.childrenFields(parentName, removeParentPrefix, level, true)
}

func (m *Metadata) StringFields(namePrefix string, fields ...map[string]Field) map[string]string {
	fs := m.Fields
	if len(fields) > 0 {
		fs = fields[0]
	}
	res := make(map[string]string)
	for k, v := range fs {
		res[fmt.Sprintf("%s%s", namePrefix, k)] = v.String()
	}
	return res
}

func (m *Metadata) FieldsFromString(ss map[string]string) map[string]Field {
	res := make(map[string]Field)
	for k, v := range ss {
		f := Field{Name: k}
		f.FromString(v)
		res[k] = f
	}
	return res
}

func (m *Metadata) UnregisterFields(fieldNames ...string) {
	for _, n := range fieldNames {
		delete(m.Fields, n)
	}
}

func (m *Metadata) clone() *Metadata {
	clone := *m
	return &clone
}

type Field struct {
	MetadataName             string `json:"metadata_name" valid:"required"`
	Name                     string `json:"name" valid:"required"`
	Type                     string `json:"type" valid:"required,in(string|int|float|bool|date|object|array)"`
	IsScalarType             bool   `json:"is_scalar_type"`
	IsRequired               bool   `json:"is_required"`
	NativeName               string `json:"native_name"`
	DisplayName              string `json:"display_name"`
	Description              string `json:"description"`
	NativeType               string `json:"native_type"`
	DefaultValue             string `json:"default_value"`
	IsAutoInc                bool   `json:"is_auto_inc"`
	IsUnique                 bool   `json:"is_unique"`
	IsPrimaryKey             bool   `json:"is_primary_key"`
	IsVirtual                bool   `json:"is_virtual"`
	ElemType                 string `json:"elem_type"`
	RelationshipKind         string `json:"rel_kind"`
	RelationshipModel        string `json:"rel_model"`
	RelationshipLocalField   string `json:"rel_local_field"`
	RelationshipForeignField string `json:"rel_foreign_field"`
	In                       string `json:"in"`
}

func (f *Field) String() string {
	var buf bytes.Buffer
	switch f.Type {
	case TypeObject:
		if f.ElemType == "" {
			buf.WriteString(f.Type)
		} else {
			buf.WriteString(f.ElemType)
		}
	case TypeArray:
		if f.ElemType == "" {
			buf.WriteString(f.Type)
		} else {
			buf.WriteString(fmt.Sprintf("[%s]", f.ElemType))
		}
	default:
		buf.WriteString(f.Type)
	}
	if f.IsRequired {
		buf.WriteRune('!')
	}
	var attrs []string
	if f.DefaultValue != "" {
		defaultValue := f.DefaultValue
		if f.Type == TypeString && defaultValue == "" {
			attrs = append(attrs, `default=""`)
		} else {
			attrs = append(attrs, fmt.Sprintf("default=%s", defaultValue))
		}
	}
	if f.NativeType != "" {
		attrs = append(attrs, fmt.Sprintf("type=%s", f.NativeType))
	}
	if f.NativeName != "" {
		attrs = append(attrs, fmt.Sprintf("native=%s", f.NativeName))
	}
	if f.IsAutoInc {
		attrs = append(attrs, "auto_inc")
	}
	if f.IsUnique {
		attrs = append(attrs, "unique")
	}
	if f.IsPrimaryKey {
		attrs = append(attrs, "primary")
	}
	if f.IsVirtual {
		attrs = append(attrs, "virtual")
	}
	if f.RelationshipKind != "" {
		attrs = append(attrs, fmt.Sprintf("rel_kind=%s", f.RelationshipKind))
	}
	if f.RelationshipModel != "" {
		attrs = append(attrs, fmt.Sprintf("rel_model=%s", f.RelationshipModel))
	}
	if f.RelationshipLocalField != "" {
		attrs = append(attrs, fmt.Sprintf("rel_local_field=%s", f.RelationshipLocalField))
	}
	if f.RelationshipForeignField != "" {
		attrs = append(attrs, fmt.Sprintf("rel_foreign_field=%s", f.RelationshipForeignField))
	}
	if f.In != "" {
		attrs = append(attrs, fmt.Sprintf("in=%s", f.In))
	}
	if len(attrs) > 0 {
		buf.WriteString(fmt.Sprintf("(%s)", strings.Join(attrs, ";")))
	}
	buf.WriteString(fmt.Sprintf(",%s", f.DisplayName))
	if f.Description != "" {
		buf.WriteString(fmt.Sprintf(",%s", f.Description))
	}
	s := buf.String()
	return s
}

func (f *Field) FromString(s string) {
	split := strings.Split(s, ",")
	if len(split) < 2 {
		return
	}
	f.DisplayName = split[1]
	if len(split) > 2 {
		f.Description = split[2]
	}
	var pt string
	si := strings.IndexByte(split[0], '(')
	if si != -1 {
		ei := strings.LastIndexByte(split[0], ')')
		for _, a := range strings.Split(split[0][si+1:ei], ";") {
			if a == "" {
				continue
			}
			prop := strings.Split(strings.TrimSpace(a), "=")
			var k, v string
			if len(prop) > 0 {
				k = strings.ToLower(prop[0])
			}
			if len(prop) > 1 {
				v = prop[1]
			}
			switch k {
			case "type":
				f.NativeType = v
			case "native":
				f.NativeName = v
			case "default":
				f.DefaultValue = v
			case "auto_inc":
				f.IsAutoInc = true
			case "unique":
				f.IsUnique = true
			case "primary":
				f.IsPrimaryKey = true
			case "virtual":
				f.IsVirtual = true
			case "rel_kind":
				f.RelationshipKind = v
			case "rel_model":
				f.RelationshipModel = v
			case "rel_local_field":
				f.RelationshipLocalField = v
			case "rel_foreign_field":
				f.RelationshipForeignField = v
			}
		}
		pt = split[0][:si]
	} else {
		pt = split[0]
	}
	var (
		rawType      string
		dataType     string
		isRequired   bool
		isScalarType = true
		elemType     string
	)
	scalarTypes := map[string]bool{
		TypeInt:    true,
		TypeFloat:  true,
		TypeString: true,
		TypeBool:   true,
		TypeTime:   true,
		TypeBlob:   true,
	}
	if strings.HasSuffix(pt, "!") {
		rawType = pt[:len(pt)-1]
		isRequired = true
	} else {
		rawType = pt
	}
	if (strings.HasPrefix(rawType, "[") && strings.HasSuffix(rawType, "]")) || rawType == TypeArray {
		dataType = TypeArray
		isScalarType = false
		if strings.HasPrefix(rawType, "[") {
			elemType = rawType[1 : len(rawType)-1]
		}
	} else if !scalarTypes[rawType] || rawType == TypeObject {
		dataType = TypeObject
		isScalarType = false
		if rawType != TypeObject {
			elemType = rawType
		}
	} else {
		dataType = rawType
	}

	f.Type = dataType
	f.IsRequired = isRequired
	f.IsScalarType = isScalarType
	f.ElemType = elemType
}

func (*Field) Metadata() Metadata {
	return Metadata{
		DisplayName: "Metadata Field",
		NativeName:  "core_field",
	}
}

func (f *Field) MustName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.NativeName
}

func (f *Field) MustNativeName() string {
	if f.NativeName != "" {
		return f.NativeName
	}
	return f.Name
}
