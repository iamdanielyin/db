package sqlite3

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/yuyitech/db/adapter/sqladapter"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"gopkg.in/guregu/null.v4"
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

// ConnectionURL implements a SQLite connection struct.
type ConnectionURL struct {
	Database string
	Options  map[string]string
}

func driverName() string {
	return "sqlite3"
}

func name(common sqlhelper.SQLCommon, ds *db.DataSource) string {
	connURL, err := ParseURL(ds.DSN)
	if err != nil {
		return ""
	}
	return connURL.Database
}

func nativeCollectionNames(common sqlhelper.SQLCommon, ds *db.DataSource) ([]string, error) {
	rows, err := common.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return nil, err
	}

	var names []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		names = append(names, tableName)
	}

	return names, nil
}

func nativeCollectionMetadata(common sqlhelper.SQLCommon, ds *db.DataSource) ([]schema.Metadata, error) {
	tableNames, err := nativeCollectionNames(common, ds)

	if err != nil {
		return nil, err
	}

	var (
		dsn      = strings.ToUpper(ds.Name)
		metadata []schema.Metadata
	)

	type column struct {
		ColumnName   string      `json:"name"`
		DataType     string      `json:"type"`
		NotNull      null.Int    `json:"notnull"`
		DefaultValue null.String `json:"dflt_value"`
		IsPrimaryKey null.Int    `json:"pk"`
	}

	for _, tableName := range tableNames {
		columnRows, err := common.Query(fmt.Sprintf("PRAGMA table_info('%s')", tableName))
		if err != nil {
			continue
		}
		var columns []column
		if err := sqlhelper.All(columnRows, &columns); err != nil {
			continue
		}
		var (
			metadataName = fmt.Sprintf("%s%s", dsn, db.ConvertMetadataName(tableName))
			fields       = make(map[string]schema.Field)
		)
		for _, item := range columns {
			field := &schema.Field{
				MetadataName: metadataName,
				Name:         item.ColumnName,
				NativeName:   item.ColumnName,
				Type:         sqlhelper.TypeMapping[item.DataType],
				NativeType:   item.DataType,
				DefaultValue: item.DefaultValue.String,
				IsRequired:   item.NotNull.Int64 == 1,
				IsPrimaryKey: item.IsPrimaryKey.Int64 == 1,
			}
			fields[field.Name] = *field
		}
		metadata = append(metadata, schema.Metadata{
			Name:           metadataName,
			NativeName:     tableName,
			DataSourceName: ds.Name,
			Fields:         fields,
		})
	}
	return metadata, nil
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

	if _, ok := conn.Options["cache"]; !ok {
		conn.Options["cache"] = "shared"
	}

	return conn, err
}
