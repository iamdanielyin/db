package schema

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultNamer(t *testing.T) {
	assert.Equal(t, DefaultNamer("User"), "users")
	assert.Equal(t, DefaultNamer("Role"), "roles")
	assert.Equal(t, DefaultNamer("RoleBinding"), "role_bindings")
	assert.Equal(t, DefaultNamer("RoleBinding", true), "role_binding")
	assert.Equal(t, DefaultNamer("Person"), "people")
}
