package ql

import (
	"fmt"
	_ "github.com/cznic/ql/driver"
	"github.com/yuyitech/db/adapter/sqladapter"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
	"net/url"
	"strings"
)

func init() {
	db.RegisterAdapter(sqladapter.NewAdapter(&sqladapter.AdapterFuncs{
		DriverName:               driverName,
		Name:                     name,
		NativeCollectionNames:    nativeCollectionNames,
		NativeCollectionMetadata: nativeCollectionMetadata,
	}))
}

const connectionScheme = `file`

type ConnectionURL struct {
	Database string
	Options  map[string]string
}

func driverName() string {
	return "ql"
}

func name(common sqlhelper.SQLCommon, ds *db.DataSource) string {
	connURL, err := ParseURL(ds.DSN)
	if err != nil {
		return ""
	}
	return connURL.Database
}

func nativeCollectionNames(common sqlhelper.SQLCommon, ds *db.DataSource) ([]string, error) {
	return nil, nil
}

func nativeCollectionMetadata(common sqlhelper.SQLCommon, ds *db.DataSource) ([]db.Metadata, error) {
	return nil, nil
}

// ParseURL parses s into a ConnectionURL struct.
func ParseURL(s string) (conn ConnectionURL, err error) {
	var u *url.URL

	if !strings.HasPrefix(s, connectionScheme+"://") {
		return conn, fmt.Errorf(`Expecting file:// connection scheme.`)
	}

	if u, err = url.Parse(s); err != nil {
		return conn, err
	}

	conn.Database = u.Host + u.Path
	conn.Options = map[string]string{}

	var vv url.Values

	if vv, err = url.ParseQuery(u.RawQuery); err != nil {
		return conn, err
	}

	for k := range vv {
		conn.Options[k] = vv.Get(k)
	}

	return conn, err
}
