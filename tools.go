package db

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"strings"
)

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
