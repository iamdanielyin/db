package db

import (
	"context"
	"database/sql"
	"github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	"github.com/yuyitech/structs"
	"gopkg.in/guregu/null.v4"
	"reflect"
	"strings"
	"sync"
	"time"
)

var (
	connMap   = make(map[string]*Connection)
	connMapMu sync.RWMutex
)

type Connection struct {
	client     Client
	cacheStore *sync.Map
}

func (c *Connection) Client() Client {
	return c.client
}

func (c *Connection) Disconnect() error {
	return c.client.Disconnect(context.Background())
}

func (c *Connection) StartTransaction() (Tx, error) {
	return c.client.StartTransaction()
}

func (c *Connection) WithTransaction(fn func(Tx) error) error {
	return c.client.WithTransaction(fn)
}

func Connect(source DataSource) (*Connection, error) {
	connMapMu.Lock()
	defer connMapMu.Unlock()

	if _, err := govalidator.ValidateStruct(&source); err != nil {
		return nil, Errorf(err.Error())
	}

	adapterMapMu.RLock()
	adapter := adapterMap[source.Adapter]
	adapterMapMu.RUnlock()

	if adapter == nil {
		return nil, Errorf(`unregistered adapter "%s"`, source.Adapter)
	}

	if _, has := connMap[source.Name]; has {
		return nil, Errorf(`data source name already exists "%s"`, source.Name)
	}

	client, err := adapter.Connect(context.Background(), source)
	if err != nil {
		return nil, err
	}
	conn := &Connection{cacheStore: &sync.Map{}}
	callbacks := callbackClientWrapper(client, conn)
	registerCreateCallbacks(callbacks)
	registerQueryCallbacks(callbacks)
	registerUpdateCallbacks(callbacks)
	registerDeleteCallbacks(callbacks)
	conn.client = callbacks
	connMap[source.Name] = conn
	return conn, nil
}

func Disconnect(names ...string) error {
	connMapMu.Lock()
	defer connMapMu.Unlock()

	if len(names) == 0 {
		for k := range connMap {
			names = append(names, k)
		}
	}

	for _, name := range names {
		name = strings.TrimSpace(name)
		if v, has := connMap[name]; has && v != nil {
			if err := v.Disconnect(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Connection) RegisterMetadata(metaOrStruct interface{}) error {
	if metaOrStruct == nil {
		return nil
	}

	var metadata Metadata
	switch v := metaOrStruct.(type) {
	case Metadata:
		metadata = v
	case *Metadata:
		metadata = *v
	default:
		parsed, err := parseStructMetadata(v)
		if err != nil {
			return Errorf("parse struct failed: %v", err)
		}
		if parsed == nil {
			return Errorf("unsupported metadata type: %v", metadata)
		}
		metadata = *parsed
	}
	metadata.source = c
	metadata.Properties = metadata.Properties.updateFieldNames()   // 先更新field.Name
	metadata.nativeProperties = metadata.Properties.nativeFields() // 再计算nativeProperties
	// 校验结构体
	if _, err := govalidator.ValidateStruct(&metadata); err != nil {
		return Errorf(err.Error())
	}

	metadataMapMu.Lock()
	metadataMap[metadata.Name] = metadata
	metadataMapMu.Unlock()

	matchLogicDeleteRules()
	matchMetadataHooks()

	return nil
}

func parseStructMetadata(v interface{}) (*Metadata, error) {
	reflectValue := reflect.Indirect(reflect.ValueOf(v))
	switch reflectValue.Kind() {
	case reflect.Struct:
		s := structs.New(v)
		metadata := Metadata{
			Name:       s.Name(),
			Properties: make(Fields),
		}
		if v, ok := v.(MetadataInterface); ok {
			vv := v.Metadata()
			if err := mergo.Merge(&metadata, vv); err != nil {
				return nil, err
			}
		}
		for _, item := range s.Fields() {
			field := parseStructFieldTag(item.Tag("db"))
			field.Name = item.Name()
			if field.Type == "" {
				switch item.Kind() {
				case reflect.Bool:
					field.Type = Bool
				case reflect.Int:
					field.Type = Int
				case reflect.Int8:
					field.Type = Int
				case reflect.Int16:
					field.Type = Int
				case reflect.Int32:
					field.Type = Int
				case reflect.Int64:
					field.Type = Int
				case reflect.Uint:
					field.Type = Int
				case reflect.Uint8:
					field.Type = Int
				case reflect.Uint16:
					field.Type = Int
				case reflect.Uint32:
					field.Type = Int
				case reflect.Uint64:
					field.Type = Int
				case reflect.Uintptr:
					field.Type = Int
				case reflect.Float32:
					field.Type = Float
				case reflect.Float64:
					field.Type = Float
				case reflect.Complex64:
					field.Type = Float
				case reflect.Complex128:
					field.Type = Float
				case reflect.Array:
					field.Type = Array
				case reflect.Map:
					field.Type = Object
				case reflect.Slice:
					field.Type = Array
				case reflect.String:
					field.Type = String
				case reflect.Struct:
					field.Type = Object
				}
				if field.Type == "" {
					itemValue := item.Value()
					switch itemValue.(type) {
					case *int, *int8, *int16, *int32, *int64,
						*uint, *uint8, *uint16, *uint32, *uint64, *uintptr,
						sql.NullInt16, *sql.NullInt16, sql.NullInt32, *sql.NullInt32, sql.NullInt64, *sql.NullInt64,
						null.Int, *null.Int:
						field.Type = Int
					case *float32, *float64, *complex64, *complex128, sql.NullFloat64, *sql.NullFloat64, null.Float, *null.Float:
						field.Type = Float
					case *bool, sql.NullBool, *sql.NullBool, null.Bool, *null.Bool:
						field.Type = Bool
					case string, *string, sql.NullString, *sql.NullString, null.String, *null.String:
						field.Type = String
					case time.Time, *time.Time, sql.NullTime, *sql.NullTime, null.Time, *null.Time:
						field.Type = Time
					}
				}
			}
			if field.Type == "" {
				continue
			}
			metadata.Properties[field.Name] = field
		}
		return &metadata, nil
	}
	return nil, nil
}

func parseStructFieldTag(tag string) (f Field) {
	for _, item := range strings.Split(tag, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		var (
			key   string
			value string
		)
		idx := strings.Index(item, "=")
		if idx > 0 {
			key = item[:idx]
			value = item[idx+1:]
		} else {
			key = item
			value = "true"
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "pk":
			f.Primary = value
		case "type":
			f.Type = value
		case "name":
			f.DisplayName = value
		case "native":
			f.NativeName = value
		case "desc":
			f.Description = value
		case "rqd":
			f.Required = value
		case "uniq":
			f.Unique = value
		case "default":
			f.DefaultValue = value
		}
	}
	return
}
