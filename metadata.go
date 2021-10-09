package db

import (
	"github.com/asaskevich/govalidator"
	"sync"
)

var (
	metadataMap   = make(map[string]map[string]Metadata)
	metadataMapMu sync.RWMutex
)

type Metadata struct {
	Name string `valid:"required,!empty"`
}

func RegisterMetadata(sourceName string, metadata Metadata) error {
	metadataMapMu.Lock()
	defer metadataMapMu.Unlock()

	if sourceName == "" {
		return Errorf("missing data source name")
	}
	if metadataMap[sourceName] == nil {
		metadataMap[sourceName] = make(map[string]Metadata)
	}
	// 校验结构体
	if _, err := govalidator.ValidateStruct(&metadata); err != nil {
		return Errorf(err.Error())
	}
	metadataMap[sourceName][metadata.Name] = metadata
	return nil
}
