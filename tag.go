package envmodel

import (
	"reflect"
	"strings"
)

const (
	TagKey = "env"
)

type TagInfo struct {
	IsDefined bool

	Key string

	Required bool
	Default  string
}

func parseTag(structTag reflect.StructTag) *TagInfo {
	var tag TagInfo
	raw := strings.TrimSpace(structTag.Get(TagKey))
	if "" == raw {
		return &TagInfo{IsDefined: false}
	}
	if "-" == raw {
		return &TagInfo{IsDefined: true, Key: "-"}
	}

	parts := strings.Split(raw, ",")

	for _, entry := range parts {
		entryParts := strings.SplitN(entry, ":", 2)
		key := entryParts[0]
		var value string
		if len(entryParts) > 1 {
			value = entryParts[1]
		}
		// TODO what about unhandled cases?
		switch key {
		case "key":
			tag.Key = value
		case "required":
			tag.Required = true
		case "default":
			tag.Default = value
		}
	}

	return &tag
}

func (t *TagInfo) IsIgnored() bool {
	return "-" == t.Key
}
