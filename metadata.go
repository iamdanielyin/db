package db

import (
	"github.com/asaskevich/govalidator"
	"strings"
	"sync"
)

var (
	metadataMap   = make(map[string]Metadata)
	metadataMapMu sync.RWMutex
)

type Metadata struct {
	source     string
	Name       string `valid:"required,!empty"`
	NativeName string
}

func (m Metadata) Source() string {
	return m.source
}

func RegisterMetadata(sourceName string, metadata Metadata) error {
	metadataMapMu.Lock()
	defer metadataMapMu.Unlock()

	if sourceName == "" {
		return Errorf("missing data source name")
	}
	if !HasSession(sourceName) {
		return Errorf(`unconnected data source "%s"`, sourceName)
	}
	metadata.source = sourceName
	// 校验结构体
	if _, err := govalidator.ValidateStruct(&metadata); err != nil {
		return Errorf(err.Error())
	}
	metadataMap[metadata.Name] = metadata
	return nil
}

func LookupMetadata(name string) (meta Metadata, err error) {
	metadataMapMu.RLock()
	defer metadataMapMu.RUnlock()

	name = strings.TrimSpace(name)
	if v, has := metadataMap[name]; !has {
		err = Errorf(`unregistered metadata "%s"`, name)
		return
	} else {
		meta = v
	}

	return
}
