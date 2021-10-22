package db

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"strings"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func IsBlankString(str string) bool {
	return govalidator.IsNull(strings.TrimSpace(str))
}

func IsNotBlankString(str string) bool {
	return !IsBlankString(str)
}

func Errorf(t string, params ...interface{}) error {
	if !strings.HasPrefix(t, "db: ") {
		t = "db: " + t
	}
	return fmt.Errorf(t, params...)
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
func JSONCopy(src, dst interface{}) (err error) {
	var data []byte
	if data, err = json.Marshal(src); err == nil {
		err = json.Unmarshal(data, dst)
	}
	return
}

func JSONStringify(value interface{}, format bool) string {
	var data []byte
	if format {
		data, _ = json.MarshalIndent(value, "", "  ")
	} else {
		data, _ = json.Marshal(value)
	}
	return string(data)
}

func JSONParse(v string, r interface{}) error {
	return json.Unmarshal([]byte(v), r)
}
