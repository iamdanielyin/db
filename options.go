package db

type InsertOptions struct {
	UpsertAssocsOnly   bool
	UpsertAssocsFields []string
}

type UpdateOptions struct {
	UpsertAssocsOnly   bool
	UpsertAssocsFields []string
}

type DeleteOptions struct {
	RemoveAssocsOnly   bool
	RemoveAssocsFields []string
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
