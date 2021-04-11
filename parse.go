package envmodel

import (
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	gatherRegexp  = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
	acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")
)

type fieldModel struct {
	Name  string
	Key   string
	Field reflect.Value
	Tag   *TagInfo
}

// Parse studies the model of obj, and reads environment variables into its fields.
func Parse(obj interface{}, option ...*Option) (*Environment, error) {
	opt := getOption(option...)

	env := NewEnvironment(opt)
	if opt.DotEnv {
		if err := LoadDotEnv(env); err != nil {
			return nil, err
		}
	}

	models, err := gatherModels(opt.Namespace, obj)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		value, ok := os.LookupEnv(model.Key)
		if !ok && "" != model.Tag.Default {
			value = model.Tag.Default
			ok = true
		}
		if !ok && model.Tag.Required {
			return nil, &ParseError{
				KeyName:   model.Key,
				FieldName: model.Name,
				TypeName:  model.Field.Type().String(),
				Reason:    ErrRequiredKeyUndefined,
			}
		}
		if "" == value {
			continue
		}

		err = parseField(value, model.Field)
		if err != nil {
			return nil, &ParseError{
				KeyName:   model.Key,
				FieldName: model.Name,
				TypeName:  model.Field.Type().String(),
				Value:     value,
				Reason:    err,
			}
		}
	}

	return env, nil
}

func gatherModels(prefix string, obj interface{}) ([]*fieldModel, error) {
	objModel := reflect.ValueOf(obj)
	if objModel.Kind() != reflect.Ptr {
		return nil, ErrInvalidTarget
	}
	objModel = objModel.Elem()
	if objModel.Kind() != reflect.Struct {
		return nil, ErrInvalidTarget
	}
	objType := objModel.Type()

	models := make([]*fieldModel, 0, objModel.NumField())
	for i := 0; i < objModel.NumField(); i++ {
		field := objModel.Field(i)
		fieldType := objType.Field(i)
		tag := parseTag(fieldType.Tag)
		if !field.CanSet() || tag.IsIgnored() {
			continue
		}

		// resolve pointers and initialise nils
		for field.Kind() == reflect.Ptr {
			if field.IsNil() {
				if field.Type().Elem().Kind() != reflect.Struct {
					break
				}
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}

		model := fieldModel{
			Name:  fieldType.Name,
			Key:   tag.Name,
			Field: field,
			Tag:   tag,
		}

		if "" == model.Key {
			words := gatherRegexp.FindAllStringSubmatch(model.Name, -1)
			var parts []string
			for _, words := range words {
				if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
					parts = append(parts, m[1], m[2])
				} else {
					parts = append(parts, words[0])
				}
			}

			model.Key = strings.ToUpper(strings.Join(parts, "_"))
		}
		if "" != prefix {
			model.Key = prefix + "_" + model.Key
		}
		models = append(models, &model)

		// flatten struct
		if field.Kind() == reflect.Struct {
			// TODO skip if decoder present
			innerPrefix := prefix
			if !fieldType.Anonymous {
				innerPrefix = model.Key
			}

			innerModels, err := gatherModels(innerPrefix, field.Addr().Interface())
			if err != nil {
				return nil, err
			}
			// replace at current node
			models = append(models[:len(models)-1], innerModels...)
			continue
		}
	}
	return models, nil
}

func parseField(value string, field reflect.Value) error {
	typ := field.Type()

	// TODO look for decoders

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var val int64
		if field.Kind() == reflect.Int64 && typ.PkgPath() == "time" && typ.Name() == "Duration" {
			d, err := time.ParseDuration(value)
			if err != nil {
				return err
			}
			val = int64(d)
		} else {
			var err error
			if val, err = strconv.ParseInt(value, 0, typ.Bits()); err != nil {
				return err
			}
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(value, 0, typ.Bits())
		if err != nil {
			return err
		}
		field.SetUint(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(value, typ.Bits())
		if err != nil {
			return err
		}
		field.SetFloat(val)
	case reflect.Slice:
		sl := reflect.MakeSlice(typ, 0, 0)
		if typ.Elem().Kind() == reflect.Uint8 {
			sl = reflect.ValueOf([]byte(value))
		} else if len(strings.TrimSpace(value)) != 0 {
			vals := strings.Split(value, ",")
			sl = reflect.MakeSlice(typ, len(vals), len(vals))
			for i, val := range vals {
				err := parseField(val, sl.Index(i))
				if err != nil {
					return err
				}
			}
		}
		field.Set(sl)
	case reflect.Map:
		mp := reflect.MakeMap(typ)
		if len(strings.TrimSpace(value)) != 0 {
			pairs := strings.Split(value, ",")
			for _, pair := range pairs {
				kvpair := strings.SplitN(pair, ":", 2)
				if len(kvpair) != 2 {
					return newTypedError(ErrInvalidMapEntry, "invalid map item: %q", pair)
				}
				k := reflect.New(typ.Key()).Elem()
				if err := parseField(kvpair[0], k); err != nil {
					return err
				}
				v := reflect.New(typ.Elem()).Elem()
				if err := parseField(kvpair[1], v); err != nil {
					return err
				}
				mp.SetMapIndex(k, v)
			}
		}
		field.Set(mp)
	default:
		// the cases above should be exhaustive
		return ErrUnsupportedType
	}

	return nil
}
