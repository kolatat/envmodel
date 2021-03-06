package envmodel

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
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
	envSet := make(map[string]string)
	for _, entry := range os.Environ() {
		kv := strings.SplitN(entry, "=", 2)
		if "" == kv[0] {
			continue
		}
		envSet[kv[0]] = kv[1]
	}

	models, err := gatherModels(opt.Namespace, obj, "", opt.Logger)
	if err != nil {
		return nil, err
	}

	for _, model := range models {
		value, ok := os.LookupEnv(model.Key)
		defaultFallback := false
		if !ok && "" != model.Tag.Default {
			value = model.Tag.Default
			ok = true
			defaultFallback = true
		}
		if !ok && model.Tag.Required {
			return nil, &ParseError{
				KeyName:   model.Key,
				FieldName: model.Name,
				TypeName:  model.Field.Type().String(),
				Reason:    ErrRequiredKeyUndefined,
			}
		}
		delete(envSet, model.Key)
		l := logEnv(model.Key, value, opt.Logger.Debug()).
			Str("fieldName", model.Name).
			Str("fieldType", model.Field.Type().String())
		if defaultFallback {
			l.Bool("defaultFallback", true)
		}
		l.Msg("parsing field")
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
	nsPrefix := opt.Namespace + " "
	for key, value := range envSet {
		if "" != opt.Namespace && !strings.HasPrefix(key, nsPrefix) {
			continue
		}
		logEnv(key, value, opt.Logger.Trace()).Msg("variable unused")
	}

	return env, nil
}

func gatherModels(keyPrefix string, obj interface{}, objectPrefix string, logger *zerolog.Logger) ([]*fieldModel, error) {
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
		if len(tag.errMsgs) > 0 {
			for _, errMsg := range tag.errMsgs {
				logger.Warn().Msgf("%s in field %s", errMsg, fieldType.Name)
			}
		}
		if !field.CanSet() || tag.IsIgnored() {
			logger.Trace().Msgf("ignoring field %s", fieldType.Name)
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
			Key:   tag.Key,
			Field: field,
			Tag:   tag,
		}
		if "" != objectPrefix {
			model.Name = objectPrefix + "." + model.Name
		}

		if "" == model.Key {
			model.Key = pascal2snake(fieldType.Name)
		}
		if "" != keyPrefix {
			model.Key = keyPrefix + "_" + model.Key
		}
		models = append(models, &model)

		// flatten struct
		if field.Kind() == reflect.Struct {
			if setterFrom(field) == nil && textUnmarshaler(field) == nil && binaryUnmarshaler(field) == nil {
				innerPrefix := keyPrefix
				if !fieldType.Anonymous {
					innerPrefix = model.Key
				}

				innerModels, err := gatherModels(innerPrefix, field.Addr().Interface(), model.Name, logger)
				if err != nil {
					return nil, err
				}
				// replace at current node
				models = append(models[:len(models)-1], innerModels...)
				continue
			}
		}
	}
	return models, nil
}

func parseField(value string, field reflect.Value) error {

	// decoder overrides
	if s := setterFrom(field); s != nil {
		return s.Set(value)
	}
	if t := textUnmarshaler(field); t != nil {
		return t.UnmarshalText([]byte(value))
	}
	if b := binaryUnmarshaler(field); b != nil {
		return b.UnmarshalBinary([]byte(value))
	}

	typ := field.Type()
	kind := typ.Kind()

	if kind == reflect.Ptr {
		typ = typ.Elem()
		if field.IsNil() {
			field.Set(reflect.New(typ))
		}
		field = field.Elem()
	}

	if err, overridden := parseOverrides(value, field); overridden {
		return err
	}

	switch kind {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(value, 0, typ.Bits())
		if err != nil {
			return err
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
