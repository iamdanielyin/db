package schema

import (
	"gopkg.in/guregu/null.v4"
	"strconv"
	"time"
)

type EnumItem struct {
	Label string
	Value string
}

func (i *EnumItem) Int() int {
	v, _ := strconv.Atoi(i.Value)
	return v
}

func (i *EnumItem) String() string {
	return i.Value
}

func (i *EnumItem) Float() float64 {
	v, _ := strconv.ParseFloat(i.Value, 64)
	return v
}

type Field struct {
	Schema            Schema
	Name              string      `db:"name" json:"name" valid:"required"`
	NativeName        string      `db:"nativeName" json:"nativeName" valid:"required"`
	Type              string      `db:"type" json:"type" valid:"required,in(string|int|float|bool|object|array|time|timestamp|password|file)"`
	NativeType        string      `db:"nativeType" json:"nativeType" valid:"required"`
	Title             string      `db:"title" json:"title" valid:"required"`
	ElementType       null.String `db:"elementType" json:"elementType"`
	Description       null.String `db:"description" json:"description"`
	Required          null.Bool   `db:"required" json:"required"`
	DefaultValue      null.String `db:"defaultValue" json:"defaultValue"`
	ForceDefaultValue null.Bool   `db:"forceDefaultValue" json:"forceDefaultValue"`
	// indexes
	Primary null.Bool   `db:"primary" json:"primary"`
	Index   null.String `db:"index" json:"index"`
	Unique  null.Bool   `db:"unique" json:"unique"`
	AutoInc null.Bool   `db:"autoInc" json:"autoInc"`
	// string
	Lowercase null.Bool   `db:"lowercase" json:"lowercase"`
	Uppercase null.Bool   `db:"uppercase" json:"uppercase"`
	Trim      null.Bool   `db:"trim" json:"trim"`
	Format    null.String `db:"format" json:"format"`
	Pattern   null.String `db:"pattern" json:"pattern"`
	Enum      []EnumItem  `db:"enum" json:"enum"`
	MinLength null.Int    `db:"minLength" json:"minLength"`
	MaxLength null.Int    `db:"maxLength" json:"maxLength"`
	// int/float
	Min          null.String `db:"min" json:"min"`
	Max          null.String `db:"max" json:"max"`
	ExclusiveMin null.Bool   `db:"exclusiveMin" json:"exclusiveMin"`
	ExclusiveMax null.Bool   `db:"exclusiveMax" json:"exclusiveMax"`
	// reference
	Ref        null.String `db:"ref" json:"ref"`
	Owner      null.Bool   `db:"owner" json:"owner"`
	LocalKey   null.String `db:"localKey" json:"localKey"`
	ForeignKey null.String `db:"foreignKey" json:"foreignKey"`
	// ui
	Label null.String `db:"label" json:"label"`
	Order null.Int    `db:"order" json:"order"`
	Group null.String `db:"group" json:"group"`
}

func (f *Field) MinInt() (int, error) {
	return strconv.Atoi(f.Min.String)
}

func (f *Field) MaxInt() (int, error) {
	return strconv.Atoi(f.Max.String)
}

func (f *Field) MinFloat() (float64, error) {
	return strconv.ParseFloat(f.Min.String, 64)
}

func (f *Field) MaxFloat() (float64, error) {
	return strconv.ParseFloat(f.Max.String, 64)
}

func (f *Field) MinTime() (time.Time, error) {
	ts, err := strconv.Atoi(f.Min.String)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(ts), 0), nil
}

func (f *Field) MaxTime() (time.Time, error) {
	ts, err := strconv.Atoi(f.Max.String)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(ts), 0), nil
}
