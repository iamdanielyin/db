package db

const (
	AssocTypeReplace = "ASSOC_REPLACE"
	AssocTypeMerge   = "ASSOC_MERGE"
	AssocTypeRemove  = "ASSOC_REMOVE"
	AssocTypeClear   = "ASSOC_CLEAR"
)

func ensureInsertAssocTypeMap(opts *InsertOptions) *InsertOptions {
	if opts != nil {
		if opts.AssocTypeMap == nil {
			opts.AssocTypeMap = make(map[string]string)
		}
	}
	return opts
}

func ensureUpdateAssocTypeMap(opts *UpdateOptions) *UpdateOptions {
	if opts != nil {
		if opts.AssocTypeMap == nil {
			opts.AssocTypeMap = make(map[string]string)
		}
	}
	return opts
}

func ensureDeleteAssocTypeMap(opts *DeleteOptions) *DeleteOptions {
	if opts != nil {
		if opts.AssocTypeMap == nil {
			opts.AssocTypeMap = make(map[string]string)
		}
	}
	return opts
}

func WithInsertOptionAssocType(fieldName string, typeValue string) func(opts *InsertOptions) {
	return func(opts *InsertOptions) {
		ensureInsertAssocTypeMap(opts).AssocTypeMap[fieldName] = typeValue
	}
}

func WithUpdateOptionAssocType(fieldName string, typeValue string) func(opts *UpdateOptions) {
	return func(opts *UpdateOptions) {
		ensureUpdateAssocTypeMap(opts).AssocTypeMap[fieldName] = typeValue
	}
}

func WithDeleteOptionAssocType(fieldName string, typeValue string) func(opts *DeleteOptions) {
	return func(opts *DeleteOptions) {
		ensureDeleteAssocTypeMap(opts).AssocTypeMap[fieldName] = typeValue
	}
}

func WithInsertOptionLooseMode(v bool) func(opts *InsertOptions) {
	return func(opts *InsertOptions) {
		ensureInsertAssocTypeMap(opts).LooseMode = v
	}
}

func WithUpdateOptionLooseMode(v bool) func(opts *UpdateOptions) {
	return func(opts *UpdateOptions) {
		ensureUpdateAssocTypeMap(opts).LooseMode = v
	}
}

func WithInsertOptionDeleteAssocs(v bool) func(opts *InsertOptions) {
	return func(opts *InsertOptions) {
		ensureInsertAssocTypeMap(opts).DeleteAssocs = v
	}
}

func WithUpdateOptionDeleteAssocs(v bool) func(opts *UpdateOptions) {
	return func(opts *UpdateOptions) {
		ensureUpdateAssocTypeMap(opts).DeleteAssocs = v
	}
}

func WithDeleteOptionDeleteAssocs(v bool) func(opts *DeleteOptions) {
	return func(opts *DeleteOptions) {
		ensureDeleteAssocTypeMap(opts).DeleteAssocs = v
	}
}

type InsertOptions struct {
	AssocTypeMap map[string]string
	LooseMode    bool
	DeleteAssocs bool
}

type UpdateOptions struct {
	AssocTypeMap map[string]string
	LooseMode    bool
	DeleteAssocs bool
}

type DeleteOptions struct {
	AssocTypeMap map[string]string
	DeleteAssocs bool
}

type PreloadOptions struct {
	Path     string
	Match    Conditional
	Select   []string
	OrderBys []string
	Page     uint
	Size     uint
}

func (opts *PreloadOptions) SetResult(res Result) Result {
	if !IsNil(opts.Match) {
		res.And(opts.Match)
	}
	if len(opts.Select) > 0 {
		res.Project(opts.Select...)
	}
	if len(opts.OrderBys) > 0 {
		res.OrderBy(opts.OrderBys...)
	}
	if opts.Size > 0 {
		res.Paginate(opts.Size)
	}
	if opts.Page > 0 {
		res.Page(opts.Page)
	}
	return res
}
