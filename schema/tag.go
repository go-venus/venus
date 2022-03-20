package schema

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/go-venus/venus/consts"
)

type Tag struct {
	Tag         reflect.StructTag
	TagSettings map[string]string
}

func (t *Tag) String() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}

func ParseTag(structTag reflect.StructTag, delimiter string) *Tag {
	t := &Tag{
		Tag:         structTag,
		TagSettings: map[string]string{},
	}
	names := strings.Split(structTag.Get(consts.OrmName), delimiter)
	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + delimiter + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			t.TagSettings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			t.TagSettings[k] = k
		}
	}

	return t
}
