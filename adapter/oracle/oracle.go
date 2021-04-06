package oracle

import (
	_ "github.com/mattn/go-oci8"
	"github.com/yuyitech/db/adapter/sqladapter"
	"github.com/yuyitech/db/adapter/sqladapter/sqlhelper"
	"github.com/yuyitech/db/pkg/db"
	"github.com/yuyitech/db/pkg/schema"
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
	return "oci8"
}

func name(common sqlhelper.SQLCommon, ds *db.DataSource) string {
	return ""
}

func nativeCollectionNames(common sqlhelper.SQLCommon, ds *db.DataSource) ([]string, error) {
	return nil, nil
}

func nativeCollectionMetadata(common sqlhelper.SQLCommon, ds *db.DataSource) ([]schema.Metadata, error) {
	return nil, nil
}
