package xtoml

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/spf13/cast"
)

type parser struct {
	tr *toml.Tree
	vv reflect.Value
	vt reflect.Type
	nf int
}

func (p *parser) checkType(i interface{}) (err error) {
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		if v.Kind() == reflect.Struct {
			p.vv = v
			p.vt = v.Type()
			p.nf = v.NumField()
			return
		}
	}
	err = errors.New("not a pointer to struct")
	return
}

func LoadConf(cf interface{}, conf string) (err error) {
	var (
		tg      string
		tgs     []string
		ok, req bool
		vv      interface{}
		fv      reflect.Value
		ft      reflect.Type
	)

	if len(conf) == 0 {
		err = errors.New("config must be set")
		return
	}
	var tr *toml.Tree
	if tr, err = toml.LoadFile(conf); err != nil {
		return
	}
	p := parser{
		tr: tr,
	}
	defer func() {
		p.tr = nil
	}()
	if err = p.checkType(cf); err != nil {
		return
	}
	for i := 0; i < p.nf; i++ {
		if tg, ok = p.vt.Field(i).Tag.Lookup("conf"); !ok {
			continue
		}
		tgs = strings.SplitN(tg, ",", 2)
		for _, t := range tgs {
			if len(t) == 0 {
				err = errors.New("invalid tag len")
				return
			}
		}
		req = false
		if len(tgs) == 2 {
			if tgs[1] == "required" {
				req = true
			} else {
				err = errors.New("invalid tag: " + tgs[1])
				return
			}
		}
		vv = p.tr.Get(tgs[0])
		if vv == nil {
			if req {
				err = errors.New("field " + tgs[0] + " must be set")
				return
			}
			continue
		}
		fv = p.vv.Field(i)
		ft = fv.Type()
		switch fv.Kind() {
		case reflect.Bool:
			var vc bool
			if vc, err = cast.ToBoolE(vv); err != nil {
				return
			}
			fv.SetBool(vc)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch fv.Interface().(type) {
			case time.Duration:
				var vc time.Duration
				if vc, err = cast.ToDurationE(vv); err != nil {
					return
				}
				fv.SetInt(int64(vc))
				// fv.Set(reflect.ValueOf(vc))
			default:
				var vc int64
				if vc, err = cast.ToInt64E(vv); err != nil {
					return
				}
				fv.SetInt(vc)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			var vc uint64
			if vc, err = cast.ToUint64E(vv); err != nil {
				return
			}
			fv.SetUint(vc)
		case reflect.Float32, reflect.Float64:
			var vc float64
			if vc, err = cast.ToFloat64E(vv); err != nil {
				return
			}
			fv.SetFloat(vc)
		case reflect.String:
			var vc string
			if vc, err = cast.ToStringE(vv); err != nil {
				return
			}
			fv.SetString(vc)
		case reflect.Slice:
			switch ft.Elem().Kind() {
			case reflect.String:
				var vc []string
				if vc, err = cast.ToStringSliceE(vv); err != nil {
					return
				}
				fv.Set(reflect.ValueOf(vc))
			default:
				err = errors.New("unsupported type: " + ft.String())
				return
			}
		case reflect.Struct:
			switch fv.Interface().(type) {
			case time.Time:
				var vc time.Time
				if vc, err = cast.ToTimeE(vv); err != nil {
					return
				}
				fv.Set(reflect.ValueOf(vc))
			default:
				err = errors.New("unsupported type: " + ft.String())
				return
			}
		default:
			err = errors.New("unsupported type: " + ft.String())
			return
		}
	}
	return
}
