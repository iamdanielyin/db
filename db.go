package db

type DataSource struct {
	Name    string `valid:"required,!empty"`
	Adapter string `valid:"required,!empty"`
	URI     string `valid:"required,!empty"`
}

func HasSession(name string) bool {
	connMapMu.RLock()
	defer connMapMu.RUnlock()

	_, has := connMap[name]
	return has
}

func Session(name string) *Connection {
	connMapMu.RLock()
	defer connMapMu.RUnlock()

	conn, has := connMap[name]
	if !has || conn == nil {
		panic(Errorf(`missing session: %s`, name))
	}
	return conn
}

func Model(name string) Collection {
	meta, err := LookupMetadata(name)
	if err != nil {
		panic(err)
	}
	return meta.Session().client.Model(meta)
}

func StartTransaction(name string) (Tx, error) {
	return Session(name).StartTransaction()
}

func WithTransaction(name string, fn func(Tx) error) error {
	return Session(name).WithTransaction(fn)
}
