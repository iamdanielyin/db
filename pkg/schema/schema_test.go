package schema_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/yuyitech/db/pkg/schema"
	"testing"
)

func TestRegisterStruct(t *testing.T) {
	if err := schema.RegisterStructs(
		&User{},
		&RoleBinding{},
		&Role{},
		&PermissionBinding{},
		&Permission{},
		&CreditCard{},
	); err != nil {
		t.Fatal(err)
	}
	userSchema, hasUserSchema := schema.Metadata("User")
	assert.Equal(t, hasUserSchema, true)
	assert.Equal(t, userSchema.Name, "User")
	assert.Equal(t, userSchema.NativeName, "users")
}
