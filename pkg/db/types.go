package db

const (
	TypeString = "string"
	TypeInt    = "int"
	TypeFloat  = "float"
	TypeBool   = "bool"
	TypeTime   = "time"
	TypeObject = "object"
	TypeArray  = "array"
	TypeBlob   = "blob"
)

func IsScalarType(t string) bool {
	return t == TypeString || t == TypeInt || t == TypeFloat || t == TypeBool || t == TypeTime
}
