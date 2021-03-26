package reflectx

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

func ToInterfaceArray(v interface{}) []interface{} {
	rv := reflect.ValueOf(v)
	switch rv.Type().Kind() {
	case reflect.Ptr:
		return ToInterfaceArray(rv.Elem().Interface())
	case reflect.Slice:
		elems := rv.Len()
		args := make([]interface{}, elems)
		for i := 0; i < elems; i++ {
			args[i] = rv.Index(i).Interface()
		}
		return args
	}
	return []interface{}{v}
}

func IsNil(i interface{}) bool {
	defer func() {
		recover()
	}()
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)
	return vi.IsNil()
}

func IsBlank(value interface{}) bool {
	if value == nil {
		return true
	}
	reflectValue := reflect.ValueOf(value)
	indirectValue := reflect.Indirect(reflectValue)
	switch indirectValue.Kind() {
	case reflect.String:
		return indirectValue.Len() == 0
	case reflect.Bool:
		return !indirectValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return indirectValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return indirectValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return indirectValue.Float() == 0
	case reflect.Map:
		return indirectValue.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return indirectValue.IsNil()
	}
	vi := reflect.ValueOf(value)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return reflect.DeepEqual(indirectValue.Interface(), reflect.Zero(indirectValue.Type()).Interface())
}

// A FieldInfo is a collection of metadata about a struct field.
type FieldInfo struct {
	Index    []int
	Path     string
	Field    reflect.StructField
	Zero     reflect.Value
	Name     string
	Options  map[string]string
	Embedded bool
	Children []*FieldInfo
	Parent   *FieldInfo
}

// A StructMap is an index of field metadata for a struct.
type StructMap struct {
	Tree  *FieldInfo
	Index []*FieldInfo
	Paths map[string]*FieldInfo
	Names map[string]*FieldInfo
}

// GetByPath returns a *FieldInfo for a given string path.
func (f StructMap) GetByPath(path string) *FieldInfo {
	return f.Paths[path]
}

// GetByTraversal returns a *FieldInfo for a given integer path.  It is
// analogous to reflect.FieldByIndex.
func (f StructMap) GetByTraversal(index []int) *FieldInfo {
	if len(index) == 0 {
		return nil
	}

	tree := f.Tree
	for _, i := range index {
		if i >= len(tree.Children) || tree.Children[i] == nil {
			return nil
		}
		tree = tree.Children[i]
	}
	return tree
}

func NewMapper(tagName string) *Mapper {
	return &Mapper{
		cache:   make(map[reflect.Type]*StructMap),
		tagName: tagName,
	}
}

// Mapper is a general purpose mapper of names to struct fields.  A Mapper
// behaves like most marshallers, optionally obeying a field tag for name
// mapping and a function to provide a basic mapping of fields to names.
type Mapper struct {
	cache      map[reflect.Type]*StructMap
	tagName    string
	tagMapFunc func(string) string
	mapFunc    func(string) string
	mutex      sync.Mutex
}

// TypeMap returns a mapping of field strings to int slices representing
// the traversal down the struct to reach the field.
func (m *Mapper) TypeMap(t reflect.Type) *StructMap {
	m.mutex.Lock()
	mapping, ok := m.cache[t]
	if !ok {
		mapping = getMapping(t, m.tagName, m.mapFunc, m.tagMapFunc)
		m.cache[t] = mapping
	}
	m.mutex.Unlock()
	return mapping
}

// FieldMap returns the mapper's mapping of field names to reflect values.  Panics
// if v's Kind is not Struct, or v is not Indirectable to a struct kind.
func (m *Mapper) FieldMap(v reflect.Value) map[string]reflect.Value {
	v = reflect.Indirect(v)
	mustBe(v, reflect.Struct)

	r := map[string]reflect.Value{}
	tm := m.TypeMap(v.Type())
	for tagName, fi := range tm.Names {
		r[tagName] = FieldByIndexes(v, fi.Index)
	}
	return r
}

// ValidFieldMap returns the mapper's mapping of field names to reflect valid
// field values.  Panics if v's Kind is not Struct, or v is not Indirectable to
// a struct kind.
func (m *Mapper) ValidFieldMap(v reflect.Value) map[string]reflect.Value {
	v = reflect.Indirect(v)
	mustBe(v, reflect.Struct)

	r := map[string]reflect.Value{}
	tm := m.TypeMap(v.Type())
	for tagName, fi := range tm.Names {
		v := ValidFieldByIndexes(v, fi.Index)
		if v.IsValid() {
			r[tagName] = v
		}
	}
	return r
}

// FieldByName returns a field by the its mapped name as a reflect.Value.
// Panics if v's Kind is not Struct or v is not Indirectable to a struct Kind.
// Returns zero Value if the name is not found.
func (m *Mapper) FieldByName(v reflect.Value, name string) reflect.Value {
	v = reflect.Indirect(v)
	mustBe(v, reflect.Struct)

	tm := m.TypeMap(v.Type())
	fi, ok := tm.Names[name]
	if !ok {
		return v
	}
	return FieldByIndexes(v, fi.Index)
}

// FieldsByName returns a slice of values corresponding to the slice of names
// for the value.  Panics if v's Kind is not Struct or v is not Indirectable
// to a struct Kind.  Returns zero Value for each name not found.
func (m *Mapper) FieldsByName(v reflect.Value, names []string) []reflect.Value {
	v = reflect.Indirect(v)
	mustBe(v, reflect.Struct)

	tm := m.TypeMap(v.Type())
	vals := make([]reflect.Value, 0, len(names))
	for _, name := range names {
		fi, ok := tm.Names[name]
		if !ok {
			vals = append(vals, *new(reflect.Value))
		} else {
			vals = append(vals, FieldByIndexes(v, fi.Index))
		}
	}
	return vals
}

// FieldByIndexesReadOnly returns a value for a particular struct traversal,
// but is not concerned with allocating nil pointers because the value is
// going to be used for reading and not setting.
func FieldByIndexesReadOnly(v reflect.Value, indexes []int) reflect.Value {
	for _, i := range indexes {
		v = reflect.Indirect(v).Field(i)
	}
	return v
}

// TraversalsByName returns a slice of int slices which represent the struct
// traversals for each mapped name.  Panics if t is not a struct or Indirectable
// to a struct.  Returns empty int slice for each name not found.
func (m *Mapper) TraversalsByName(t reflect.Type, names []string) [][]int {
	t = Deref(t)
	mustBe(t, reflect.Struct)
	tm := m.TypeMap(t)

	r := make([][]int, 0, len(names))
	for _, name := range names {
		fi, ok := tm.Names[name]
		if !ok {
			r = append(r, []int{})
		} else {
			r = append(r, fi.Index)
		}
	}
	return r
}

// FieldByIndexes returns a value for a particular struct traversal.
func FieldByIndexes(v reflect.Value, indexes []int) reflect.Value {

	for _, i := range indexes {
		v = reflect.Indirect(v).Field(i)
		// if this is a pointer, it's possible it is nil
		if v.Kind() == reflect.Ptr && v.IsNil() {
			alloc := reflect.New(Deref(v.Type()))
			v.Set(alloc)
		}
		if v.Kind() == reflect.Map && v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
	}

	return v
}

// ValidFieldByIndexes returns a value for a particular struct traversal.
func ValidFieldByIndexes(v reflect.Value, indexes []int) reflect.Value {

	for _, i := range indexes {
		v = reflect.Indirect(v)
		if !v.IsValid() {
			return reflect.Value{}
		}
		v = v.Field(i)
		// if this is a pointer, it's possible it is nil
		if (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map) && v.IsNil() {
			return reflect.Value{}
		}
	}

	return v
}

// Deref is Indirect for reflect.Types
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// -- helpers & utilities --

type kinder interface {
	Kind() reflect.Kind
}

// mustBe checks a value against a kind, panicing with a reflect.ValueError
// if the kind isn't that which is required.
func mustBe(v kinder, expected reflect.Kind) {
	k := v.Kind()
	if k != expected {
		panic(&reflect.ValueError{Method: methodName(), Kind: k})
	}
}

// methodName is returns the caller of the function calling methodName
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

type typeQueue struct {
	t  reflect.Type
	fi *FieldInfo
	pp string // Parent path
}

// A copying append that creates a new slice each time.
func apnd(is []int, i int) []int {
	x := make([]int, len(is)+1)
	copy(x, is)
	x[len(x)-1] = i
	return x
}

// getMapping returns a mapping for the t type, using the tagName, mapFunc and
// tagMapFunc to determine the canonical names of fields.
func getMapping(t reflect.Type, tagName string, mapFunc, tagMapFunc func(string) string) *StructMap {
	if tagName == "" {
		tagName = "json"
	}
	var m []*FieldInfo

	root := &FieldInfo{}
	var queue []typeQueue
	queue = append(queue, typeQueue{Deref(t), root, ""})

	for len(queue) != 0 {
		// pop the first item off of the queue
		tq := queue[0]
		queue = queue[1:]
		nChildren := 0
		if tq.t.Kind() == reflect.Struct {
			nChildren = tq.t.NumField()
		}
		tq.fi.Children = make([]*FieldInfo, nChildren)

		// iterate through all of its fields
		for fieldPos := 0; fieldPos < nChildren; fieldPos++ {
			f := tq.t.Field(fieldPos)

			fi := FieldInfo{}
			fi.Field = f
			fi.Zero = reflect.New(f.Type).Elem()
			fi.Options = map[string]string{}

			var tag, name string
			if tagName != "" && strings.Contains(string(f.Tag), tagName+":") {
				tag = f.Tag.Get(tagName)
				if tagName == "json" {
					split := strings.Split(tag, ",")
					tag = split[0]
				}
				name = tag
			} else {
				if mapFunc != nil {
					name = mapFunc(f.Name)
				}
			}

			parts := strings.Split(name, ",")
			if len(parts) > 1 {
				name = parts[0]
				for _, opt := range parts[1:] {
					kv := strings.Split(opt, "=")
					if len(kv) > 1 {
						fi.Options[kv[0]] = kv[1]
					} else {
						fi.Options[kv[0]] = ""
					}
				}
			}

			if tagMapFunc != nil {
				tag = tagMapFunc(tag)
			}

			fi.Name = name

			if tq.pp == "" || (tq.pp == "" && tag == "") {
				fi.Path = fi.Name
			} else {
				fi.Path = fmt.Sprintf("%s.%s", tq.pp, fi.Name)
			}

			// if the name is "-", disabled via a tag, skip it
			if name == "-" {
				continue
			}

			// skip unexported fields
			if len(f.PkgPath) != 0 && !f.Anonymous {
				continue
			}

			// bfs search of anonymous embedded structs
			if f.Anonymous {
				pp := tq.pp
				if tag != "" {
					pp = fi.Path
				}

				fi.Embedded = true
				fi.Index = apnd(tq.fi.Index, fieldPos)
				nChildren := 0
				ft := Deref(f.Type)
				if ft.Kind() == reflect.Struct {
					nChildren = ft.NumField()
				}
				fi.Children = make([]*FieldInfo, nChildren)
				queue = append(queue, typeQueue{Deref(f.Type), &fi, pp})
			} else if fi.Zero.Kind() == reflect.Struct || (fi.Zero.Kind() == reflect.Ptr && fi.Zero.Type().Elem().Kind() == reflect.Struct) {
				fi.Index = apnd(tq.fi.Index, fieldPos)
				fi.Children = make([]*FieldInfo, Deref(f.Type).NumField())
				queue = append(queue, typeQueue{Deref(f.Type), &fi, fi.Path})
			}

			fi.Index = apnd(tq.fi.Index, fieldPos)
			fi.Parent = tq.fi
			tq.fi.Children[fieldPos] = &fi
			m = append(m, &fi)
		}
	}

	flds := &StructMap{Index: m, Tree: root, Paths: map[string]*FieldInfo{}, Names: map[string]*FieldInfo{}}
	for _, fi := range flds.Index {
		flds.Paths[fi.Path] = fi
		if fi.Name != "" && !fi.Embedded {
			flds.Names[fi.Path] = fi
		}
	}

	return flds
}
