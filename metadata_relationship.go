package db

const (
	RelationshipHasOne  = "HAS_ONE"
	RelationshipHasMany = "HAS_MANY"
	RelationshipRefOne  = "REF_ONE"
	RelationshipRefMany = "REF_MANY"
)

type Relationship struct {
	Type                     string
	SrcFieldName             string
	DstFieldName             string
	MetadataName             string
	IntermediateMetadataName string
	IntermediateSrcFieldName string
	IntermediateDstFieldName string
}
