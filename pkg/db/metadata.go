package db

import (
	"bytes"
	"github.com/asaskevich/govalidator"
	"github.com/imdario/mergo"
	"github.com/yuyitech/db/internal/safe_map"
	"github.com/yuyitech/db/pkg/logger"
	"github.com/yuyitech/db/pkg/schema"
	"gopkg.in/guregu/null.v4"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	metadata        = make(map[string]schema.Metadata)
	metadataMu      sync.RWMutex
	metaNamingCache = safe_map.NewSafeMapString()
)

type IMetadata interface {
	Metadata() schema.Metadata
}

func ConvertMetadataName(name string) string {
	if v := metaNamingCache.Get(name); v != "" {
		return v
	}

	if name == "" {
		return ""
	}

	var (
		b     bytes.Buffer
		width int
	)

	for i, r := range name {
		if r == '_' || r == '-' || r == '=' || r == ' ' ||
			(r <= 0x7F && ('0' <= r && r <= '9')) {
			width++
			continue
		}
		if width > 0 || i == 0 || ('A' <= r && r <= 'Z') {
			b.WriteRune(unicode.ToUpper(r))
			width = 0
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}

	s := b.String()
	metaNamingCache.Set(name, s)

	return s
}

func RegisterModel(v interface{}) error {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	var meta schema.Metadata
	if a, ok := v.(schema.Metadata); ok {
		meta = a
	} else if a, ok := v.(*schema.Metadata); ok {
		meta = *a
	} else {
		v, err := ParseMetadata(v)
		if err != nil {
			return err
		}
		meta = *v
	}
	if meta.Name == "" {
		return nil
	}

	if _, err := govalidator.ValidateStruct(&meta); err != nil {
		return err
	}
	if meta.NativeName == "" {
		meta.NativeName = meta.Name
	}
	if meta.DataSourceName == "" {
		meta.DataSourceName = defaultDataSourceName
	}

	metadata[meta.Name] = meta
	return nil
}

func UnregisterModel(name string) {
	metadataMu.Lock()
	defer metadataMu.Unlock()

	if _, has := metadata[name]; has {
		delete(metadata, name)
	}
}

func Meta(name string) (schema.Metadata, bool) {
	metadataMu.RLock()
	defer metadataMu.RUnlock()

	meta, has := metadata[name]
	return meta, has
}

func Model(name string) Collection {
	meta, has := Meta(name)
	if !has || meta.Name == "" {
		return nil
	}

	d := Session(meta.DataSourceName)
	if d == nil {
		dsn := meta.DataSourceName
		if dsn == "" {
			dsn = "default"
		}
		logger.ERROR("Data source '%s' is not registered", dsn)
		return nil
	}
	m := d.Model(name)
	if m == nil {
		logger.ERROR("Model '%s' is not registered", name)
	}
	return m
}
