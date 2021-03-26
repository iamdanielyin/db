package sqlhelper

import (
	"github.com/yuyitech/db/pkg/db"
)

var TypeMapping = map[string]string{
	"int":                      db.TypeInt,
	"integer":                  db.TypeInt,
	"tinyint":                  db.TypeInt,
	"smallint":                 db.TypeInt,
	"mediumint":                db.TypeInt,
	"bigint":                   db.TypeInt,
	"int unsigned":             db.TypeInt,
	"integer unsigned":         db.TypeInt,
	"tinyint unsigned":         db.TypeInt,
	"smallint unsigned":        db.TypeInt,
	"mediumint unsigned":       db.TypeInt,
	"bigint unsigned":          db.TypeInt,
	"bit":                      db.TypeInt,
	"bool":                     db.TypeBool,
	"boolean":                  db.TypeBool,
	"enum":                     db.TypeString,
	"set":                      db.TypeString,
	"varchar":                  db.TypeString,
	"char":                     db.TypeString,
	"tinytext":                 db.TypeString,
	"mediumtext":               db.TypeString,
	"text":                     db.TypeString,
	"longtext":                 db.TypeString,
	"blob":                     db.TypeString,
	"tinyblob":                 db.TypeString,
	"mediumblob":               db.TypeString,
	"longblob":                 db.TypeString,
	"date":                     db.TypeTime,
	"datetime":                 db.TypeTime,
	"timestamp":                db.TypeTime,
	"timestamp with time zone": db.TypeTime,
	"time":                     db.TypeTime,
	"float":                    db.TypeFloat,
	"double":                   db.TypeFloat,
	"decimal":                  db.TypeFloat,
	"binary":                   db.TypeString,
	"varbinary":                db.TypeString,

	"NUMBER":  db.TypeFloat,
	"INTEGER": db.TypeInt,
	"TEXT":    db.TypeString,
}
