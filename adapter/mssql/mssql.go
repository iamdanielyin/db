package mssql

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
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
	return "mssql"
}

func name(common sqlhelper.SQLCommon, _ *db.DataSource) string {
	row := common.QueryRow("SELECT DB_NAME() AS name")

	var dbName string
	_ = row.Scan(&dbName)
	return dbName
}

func nativeCollectionNames(common sqlhelper.SQLCommon, _ *db.DataSource) ([]string, error) {
	rows, err := common.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_CATALOG = DB_NAME()")
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
	type table struct {
		TableName    string `json:"TABLE_NAME"`
		TableComment string `json:"TABLE_COMMENT"`
	}
	tableRows, err := common.Query("SELECT TABLE_NAME FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_CATALOG = DB_NAME()")
	if err != nil {
		return nil, err
	}
	var tables []table
	if err := sqlhelper.All(tableRows, &tables); err != nil {
		return nil, err
	}

	type column struct {
		TableName     string      `json:"TABLE_NAME"`
		ColumnName    string      `json:"COLUMN_NAME"`
		DataType      string      `json:"DATA_TYPE"`
		ColumnDefault null.String `json:"COLUMN_DEFAULT"`
		IsNullable    null.String `json:"IS_NULLABLE"`
		ColumnComment null.String `json:"COLUMN_COMMENT"`
		ColumnKey     null.String `json:"COLUMN_KEY"`
		Extra         null.String `json:"EXTRA"`
	}
	columnRows, err := common.Query("SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_CATALOG = DB_NAME()")
	if err != nil {
		return nil, err
	}
	var columns []column
	if err := sqlhelper.All(columnRows, &columns); err != nil {
		return nil, err
	}
	fieldMap := make(map[string]map[string]schema.Field)
	for _, item := range columns {
		k := strings.ToUpper(item.ColumnKey.String)
		f := schema.Field{
			Name:         item.ColumnName,
			NativeName:   item.ColumnName,
			Description:  item.ColumnComment.String,
			Type:         sqlhelper.TypeMapping[item.DataType],
			NativeType:   item.DataType,
			DefaultValue: item.ColumnDefault.String,
			IsRequired:   strings.ToUpper(item.IsNullable.String) == "NO",
			IsAutoInc:    strings.Contains(strings.ToUpper(item.Extra.String), "AUTO_INCREMENT"),
			IsUnique:     strings.Contains(k, "UNI"),
			IsPrimaryKey: strings.Contains(k, "PRI"),
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
	for _, item := range tables {
		name := fmt.Sprintf("%s%s", dsn, db.ConvertMetadataName(item.TableName))
		for key, value := range fieldMap[item.TableName] {
			value.MetadataName = name
			fieldMap[item.TableName][key] = value
		}
		meta := schema.Metadata{
			Name:           name,
			NativeName:     item.TableName,
			DisplayName:    item.TableComment,
			DataSourceName: ds.Name,
			Fields:         fieldMap[item.TableName],
		}
		metadata = append(metadata, meta)
	}
	return metadata, nil
}
