package db

type DataSource struct {
	Name    string `valid:"required,!empty"`
	Adapter string `valid:"required,!empty"`
	URI     string `valid:"required,!empty"`
}

func LookupSession(name string) (*Connection, bool) {
	connMapMu.RLock()
	defer connMapMu.RUnlock()

	conn, has := connMap[name]
	return conn, has
}

func Session(name string) *Connection {
	conn, has := LookupSession(name)
	if !has {
		panic(Errorf(`missing session: %s`, name))
	}
	return conn
}

func Raw(name string, raw string, values ...interface{}) error {
	conn, has := LookupSession(name)
	if has {
		return conn.Raw(raw, values...)
	}
	return Errorf(`missing session: %s`, name)
}

func HasModel(name string) bool {
	_, err := LookupMetadata(name)
	return err == nil
}

func Model(name string) Collection {
	meta, err := LookupMetadata(name)
	if err != nil {
		meta = Metadata{Name: name}
	}
	return meta.Session().Client().Model(meta)
}

func StartTransaction(name string) (Tx, error) {
	return Session(name).StartTransaction()
}

func WithTransaction(name string, fn func(Tx) error) error {
	return Session(name).WithTransaction(fn)
}
