package db

import (
	"github.com/asaskevich/govalidator"
)

func init() {
	govalidator.TagMap["empty"] = IsBlankString
	govalidator.TagMap["!empty"] = IsNotBlankString
}
