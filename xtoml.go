package xtoml

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cast"
)

const (
	tagName = "conf"
	reqName = "required"
)

type confParser struct {
	tag string        // tag used for config
	vv  reflect.Value // conf struct value
	nf  int           // number fields in struct
	tt  *toml.Tree    // conf tree
}

type fieldTag struct {
	tag string // struct field tag
	req bool   // is field required
}

func (f *fieldTag) required() bool {
	return f.req
}

func (f *fieldTag) getTag() string {
	return f.tag
}

func newParser(cf interface{}, conf, tag string) (*confParser, error) {
	if len(conf) == 0 {
		return nil, fmt.Errorf("config must be set")
	}
	v := reflect.ValueOf(cf)
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("argument not a pointer")
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("argument not a pointer to struct")
	}
	tt, err := toml.LoadFile(conf)
	if err != nil {
		return nil, err
	}
	if len(tag) == 0 {
		tag = tagName
	}
	return &confParser{
		tag: tag,
		vv:  v,
		nf:  v.NumField(),
		tt:  tt,
	}, nil
}

func (c *confParser) len() int {
	return c.nf
}

func (c *confParser) getFieldTag(n int) (*fieldTag, error) {
	tg, ok := c.vv.Type().Field(n).Tag.Lookup(c.tag)
	if !ok {
		return nil, nil
	}
	tgs := strings.Split(tg, ",")
	if len(tgs) < 1 {
		return nil, fmt.Errorf("invalid tag: %s", tg)
	}
	tt := &fieldTag{
		tag: tgs[0],
	}
	for i := 1; i < len(tgs); i++ {
		switch tgs[i] {
		case reqName:
			tt.req = true
		default:
			return nil, fmt.Errorf("invalid tag: %s", tgs[i])
		}
	}
	return tt, nil
}

func (c *confParser) getField(n int) reflect.Value {
	return c.vv.Field(n)
}

func (c *confParser) getTomlData(n int) (interface{}, *fieldTag, error) {
	tt, err := c.getFieldTag(n)
	if err != nil { // error in tag
		return nil, nil, err
	}
	if tt == nil { // no this tag
		return nil, nil, nil
	}
	if v := c.tt.Get(tt.getTag()); v != nil {
		return v, tt, nil
	}
	if tt.required() {
		return nil, nil, fmt.Errorf("field %s must be set", tt.getTag())
	}
	return nil, nil, nil
}

func LoadConfExt(cf interface{}, conf, tag string) error {
	p, err := newParser(cf, conf, tag)
	if err != nil {
		return err
	}
	for i := 0; i < p.len(); i++ {
		vv, tt, err := p.getTomlData(i)
		if err != nil {
			return err
		}
		if vv == nil {
			continue
		}
		fv := p.getField(i)
		switch fv.Kind() {
		case reflect.Bool:
			if tmp, err := cast.ToBoolE(vv); err != nil {
				return err
			} else {
				fv.SetBool(tmp)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			var tmp int64
			switch fv.Interface().(type) {
			case time.Duration:
				if td, err := cast.ToDurationE(vv); err != nil {
					return err
				} else {
					tmp = int64(td)
				}
			default:
				if tmp, err = cast.ToInt64E(vv); err != nil {
					return err
				}
			}
			fv.SetInt(tmp)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if tmp, err := cast.ToUint64E(vv); err != nil {
				return err
			} else {
				fv.SetUint(tmp)
			}
		case reflect.Float32, reflect.Float64:
			if tmp, err := cast.ToFloat64E(vv); err != nil {
				return err
			} else {
				fv.SetFloat(tmp)
			}
		case reflect.String:
			if tmp, err := cast.ToStringE(vv); err != nil {
				return err
			} else {
				fv.SetString(tmp)
			}
		case reflect.Slice:
			switch fv.Type().Elem().Kind() {
			case reflect.String:
				if tmp, err := cast.ToStringSliceE(vv); err != nil {
					return err
				} else if tt.required() && len(tmp) == 0 {
					return fmt.Errorf("field %s must be set", tt.getTag())
				} else {
					fv.Set(reflect.ValueOf(tmp))
				}
			default:
				return fmt.Errorf("unsupported type: %s", fv.Type())
			}
		case reflect.Struct:
			switch fv.Interface().(type) {
			case time.Time:
				if tmp, err := cast.ToTimeE(vv); err != nil {
					return err
				} else {
					fv.Set(reflect.ValueOf(tmp))
				}
			default:
				return fmt.Errorf("unsupported type: %s", fv.Type())
			}
		default:
			return fmt.Errorf("unsupported type: %s", fv.Type())
		}
	}
	return nil
}

func LoadConf(cf interface{}, conf string) error {
	return LoadConfExt(cf, conf, tagName)
}
