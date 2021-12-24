package db

const (
	AssociationActionReplace = "REPLACE"
	AssociationActionMerge   = "MERGE"
	AssociationActionRemove  = "REMOVE"
)

type AssociationOption struct {
	Field         string
	Action        string
	AutoUpdate    bool
	AutoCreate    bool
	AutoRemove    bool
	DeleteObjects bool
}

type InsertOptions struct {
}

type UpdateOptions struct {
	AssocOptions []*AssociationOption
}

type DeleteOptions struct {
	AssocOptions []*AssociationOption
}

type PreloadOptions struct {
	Path     string
	Match    interface{}
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
