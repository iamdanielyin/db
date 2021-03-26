package sqlhelper

import (
	"errors"
	"fmt"
	"github.com/yuyitech/db/internal/reflectx"
	"reflect"
	"sort"
)

type MapOptions struct {
	IncludeZeroed bool
	IncludeNil    bool
}

var defaultMapOptions = MapOptions{
	IncludeZeroed: false,
	IncludeNil:    false,
}

type fieldValue struct {
	fields []string
	values []interface{}
}

func (fv *fieldValue) Len() int {
	return len(fv.fields)
}

func (fv *fieldValue) Swap(i, j int) {
	fv.fields[i], fv.fields[j] = fv.fields[j], fv.fields[i]
	fv.values[i], fv.values[j] = fv.values[j], fv.values[i]
}

func (fv *fieldValue) Less(i, j int) bool {
	return fv.fields[i] < fv.fields[j]
}

type hasIsZero interface {
	IsZero() bool
}

func Map(item interface{}, options *MapOptions) ([]string, []interface{}, error) {
	var fv fieldValue
	if options == nil {
		options = &defaultMapOptions
	}

	itemV := reflect.ValueOf(item)
	if !itemV.IsValid() {
		return nil, nil, nil
	}

	itemT := itemV.Type()

	if itemT.Kind() == reflect.Ptr {
		// Single dereference. Just in case the user passes a pointer to struct
		// instead of a struct.
		item = itemV.Elem().Interface()
		itemV = reflect.ValueOf(item)
		itemT = itemV.Type()
	}

	switch itemT.Kind() {
	case reflect.Struct:
		fieldMap := mapper.TypeMap(itemT).Names
		nfields := len(fieldMap)

		fv.values = make([]interface{}, 0, nfields)
		fv.fields = make([]string, 0, nfields)

		for _, fi := range fieldMap {

			// Field options
			_, tagOmitEmpty := fi.Options["omitempty"]

			fld := reflectx.FieldByIndexesReadOnly(itemV, fi.Index)
			if fld.Kind() == reflect.Ptr && fld.IsNil() {
				if tagOmitEmpty && !options.IncludeNil {
					continue
				}
				fv.fields = append(fv.fields, fi.Name)
				fv.values = append(fv.values, nil)
				continue
			}

			value := fld.Interface()

			isZero := false
			if t, ok := fld.Interface().(hasIsZero); ok {
				if t.IsZero() {
					isZero = true
				}
			} else if fld.Kind() == reflect.Array || fld.Kind() == reflect.Slice {
				if fld.Len() == 0 {
					isZero = true
				}
			} else if reflect.DeepEqual(fi.Zero.Interface(), value) {
				isZero = true
			}

			if isZero && !options.IncludeZeroed {
				continue
			}

			fv.fields = append(fv.fields, fi.Name)
			fv.values = append(fv.values, value)
		}

	case reflect.Map:
		nfields := itemV.Len()
		fv.values = make([]interface{}, nfields)
		fv.fields = make([]string, nfields)
		mkeys := itemV.MapKeys()

		for i, keyV := range mkeys {
			valv := itemV.MapIndex(keyV)
			fv.fields[i] = fmt.Sprintf("%v", keyV.Interface())

			v := valv.Interface()

			fv.values[i] = v
		}
	default:
		return nil, nil, errors.New(`db: expecting a pointer to either a map or a struct`)
	}

	sort.Sort(&fv)

	return fv.fields, fv.values, nil
}
