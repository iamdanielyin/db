package db

import (
	"github.com/asaskevich/govalidator"
	"testing"
)

type user struct {
	FirstName string `valid:"required,!empty"`
	LastName  string `valid:"empty"`
}

func TestIsBlankString(t *testing.T) {
	var a1 = user{FirstName: "Foo", LastName: "Bar"}
	if _, err := govalidator.ValidateStruct(a1); err == nil {
		t.Fatal("LastName does validate as empty.")
	}
	var a2 = user{FirstName: "Foo", LastName: ""}
	if _, err := govalidator.ValidateStruct(a2); err != nil {
		t.Fatal("LastName does validate as empty.")
	}
}

func TestIsNotBlankString(t *testing.T) {
	var a1 = user{FirstName: " "}
	if _, err := govalidator.ValidateStruct(a1); err == nil {
		t.Fatal("FirstName does validate as not empty.")
	}
	var a2 = user{FirstName: "Foo"}
	if _, err := govalidator.ValidateStruct(a2); err != nil {
		t.Fatal("FirstName does validate as not empty.")
	}
}
