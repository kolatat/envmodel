package envmodel

import (
	"reflect"
	"time"
)

func parseOverrides(value string, field reflect.Value) (error, bool) {
	typ := field.Type()
	kind := typ.Kind()

	if kind == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
		d, err := time.ParseDuration(value)
		if err != nil {
			return err, true
		}
		field.SetInt(int64(d))
		return nil, true
	}

	return nil, false
}
