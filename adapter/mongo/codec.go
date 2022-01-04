package mongo

import (
	"github.com/iancoleman/strcase"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"reflect"
	"strings"
)

var structTagParser bsoncodec.StructTagParserFunc = func(sf reflect.StructField) (bsoncodec.StructTags, error) {
	key := strcase.ToSnake(sf.Name)
	tag, ok := sf.Tag.Lookup("bson")
	if !ok {
		tag, ok = sf.Tag.Lookup("json")
	}
	if !ok && !strings.Contains(string(sf.Tag), ":") && len(sf.Tag) > 0 {
		tag = string(sf.Tag)
	}

	return parseTags(key, tag)
}

func parseTags(key string, tag string) (bsoncodec.StructTags, error) {
	var st bsoncodec.StructTags
	if tag == "-" {
		st.Skip = true
		return st, nil
	}

	for idx, str := range strings.Split(tag, ",") {
		if idx == 0 && str != "" {
			key = str
		}
		switch str {
		case "omitempty":
			st.OmitEmpty = true
		case "minsize":
			st.MinSize = true
		case "truncate":
			st.Truncate = true
		case "inline":
			st.Inline = true
		}
	}

	st.Name = key

	return st, nil
}
