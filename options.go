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
