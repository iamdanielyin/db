package postgres

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/yuyitech/db/adapter/sqladapter"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
	"gopkg.in/guregu/null.v4"
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

func driverName() string {
	return "postgres"
}

func name(common sqlhelper.SQLCommon, ds *db.DataSource) string {
	row := common.QueryRow("SELECT current_database() AS name")

	var dbName string
	_ = row.Scan(&dbName)
	return dbName
}

func nativeCollectionNames(common sqlhelper.SQLCommon, ds *db.DataSource) ([]string, error) {
	rows, err := common.Query(`
		SELECT table_name
		FROM information_schema.tables
		WHERE table_catalog = current_database()
		  AND table_schema = 'public'
		ORDER BY table_name`)
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
	tables, err := nativeCollectionNames(common, ds)
	if err != nil {
		return nil, err
	}

	type column struct {
		TableName      string      `json:"table_name"`
		ColumnName     string      `json:"column_name"`
		DataType       string      `json:"data_type"`
		ColumnDefault  null.String `json:"column_default"`
		IsNullable     null.String `json:"is_nullable"`
		ConstraintType null.String `json:"constraint_type"`
	}
	columnRows, err := common.Query(`
		SELECT a.table_name, a.column_name, a.data_type, a.column_default, a.is_nullable, b.constraint_type
		FROM (
				 SELECT *, concat(table_name, '.', column_name) column_key
				 FROM information_schema.columns
				 WHERE table_catalog = current_database()
				   and table_schema = 'public'
			 ) as a
				 LEFT JOIN (
			SELECT a.table_name, a.column_name, b.constraint_type, concat(a.table_name, '.', a.column_name) column_key
			FROM information_schema.constraint_column_usage a
					 LEFT JOIN
				 information_schema.table_constraints b ON a.constraint_name = b.constraint_name
			WHERE a.table_catalog = current_database()
			  AND a.table_schema = 'public'
		) b ON a.column_key = b.column_key`)
	if err != nil {
		return nil, err
	}
	var columns []column
	if err := sqlhelper.All(columnRows, &columns); err != nil {
		return nil, err
	}
	fieldMap := make(map[string]map[string]schema.Field)
	for _, item := range columns {
		k := strings.ToUpper(item.ConstraintType.String)
		f := schema.Field{
			Name:         item.ColumnName,
			NativeName:   item.ColumnName,
			Type:         sqlhelper.TypeMapping[item.DataType],
			NativeType:   item.DataType,
			DefaultValue: item.ColumnDefault.String,
			IsAutoInc:    item.DataType == "integer" && strings.HasPrefix(item.ColumnDefault.String, "nextval("),
			IsRequired:   strings.ToUpper(item.IsNullable.String) == "YES",
			IsUnique:     k == "UNIQUE",
			IsPrimaryKey: k == "PRIMARY KEY",
		}

		v := fieldMap[item.TableName]
		if v == nil {
			v = make(map[string]schema.Field)
		}
		v[f.Name] = f

		fieldMap[item.TableName] = v
	}
	dsn := strings.ToUpper(ds.Name)
	var metadata []schema.Metadata
	for _, tableName := range tables {
		name := fmt.Sprintf("%s%s", dsn, db.ConvertMetadataName(tableName))
		for key, value := range fieldMap[tableName] {
			value.MetadataName = name
			fieldMap[tableName][key] = value
		}
		meta := schema.Metadata{
			Name:           name,
			NativeName:     tableName,
			DataSourceName: ds.Name,
			Fields:         fieldMap[tableName],
		}
		metadata = append(metadata, meta)
	}
	return metadata, nil
}
