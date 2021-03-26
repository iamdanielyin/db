package mysql

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yuyitech/db/adapter/sqladapter"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
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
	return "mysql"
}

func name(common sqlhelper.SQLCommon, _ *db.DataSource) string {
	row := common.QueryRow("SELECT DATABASE() AS name")

	var dbName string
	_ = row.Scan(&dbName)
	return dbName
}

func nativeCollectionNames(common sqlhelper.SQLCommon, _ *db.DataSource) ([]string, error) {
	rows, err := common.Query("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()")
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

func nativeCollectionMetadata(common sqlhelper.SQLCommon, ds *db.DataSource) ([]db.Metadata, error) {
	type table struct {
		TableName    string `json:"TABLE_NAME"`
		TableComment string `json:"TABLE_COMMENT"`
	}
	tableRows, err := common.Query("SELECT TABLE_NAME, TABLE_COMMENT FROM information_schema.TABLES WHERE table_schema = DATABASE()")
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
	columnRows, err := common.Query("SELECT * FROM information_schema.COLUMNS WHERE table_schema = DATABASE()")
	if err != nil {
		return nil, err
	}
	var columns []column
	if err := sqlhelper.All(columnRows, &columns); err != nil {
		return nil, err
	}
	fieldMap := make(map[string]map[string]db.Field)
	for _, item := range columns {
		k := strings.ToUpper(item.ColumnKey.String)
		f := db.Field{
			Name:         item.ColumnName,
			NativeName:   item.ColumnName,
			Description:  item.ColumnComment.String,
			Type:         sqlhelper.TypeMapping[item.DataType],
			NativeType:   item.DataType,
			DefaultValue: item.ColumnDefault.String,
			IsRequired:   strings.ToUpper(item.IsNullable.String) == "YES",
			IsAutoInc:    strings.Contains(strings.ToUpper(item.Extra.String), "AUTO_INCREMENT"),
			IsUnique:     strings.Contains(k, "UNI"),
			IsPrimaryKey: strings.Contains(k, "PRI"),
		}

		v := fieldMap[item.TableName]
		if v == nil {
			v = make(map[string]db.Field)
		}
		v[f.Name] = f

		fieldMap[item.TableName] = v
	}
	dsn := strings.ToUpper(ds.Name)
	var metadata []db.Metadata
	for _, item := range tables {
		name := fmt.Sprintf("%s%s", dsn, db.ConvertMetadataName(item.TableName))
		for key, value := range fieldMap[item.TableName] {
			value.MetadataName = name
			fieldMap[item.TableName][key] = value
		}
		meta := db.Metadata{
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
