package grape

import (
	"reflect"
)

// H represents presentation options for conditional fields, equivalent to map[string]any
type H map[string]any

type EntityField struct {
	Name      string
	Presenter *Entity
	Func      func(any) any
	JSONKey   string
	Condition func(any, H) bool
	Default   any
	Desc      string
	Example   any
}

type Entity struct {
	Fields []*EntityField
}

func NewEntity() *Entity { return &Entity{Fields: []*EntityField{}} }

func (p *Entity) Field(name string) *EntityField {
	pf := &EntityField{Name: name, JSONKey: name}
	p.Fields = append(p.Fields, pf)
	return pf
}

func (pf *EntityField) As(jsonKey string) *EntityField    { pf.JSONKey = jsonKey; return pf }
func (pf *EntityField) WithSchema(p *Entity) *EntityField { pf.Presenter = p; return pf }
func (pf *EntityField) FieldFunc(f func(any) any) *EntityField {
	pf.Func = f
	return pf
}
func (pf *EntityField) If(cond func(any, H) bool) *EntityField {
	pf.Condition = cond
	return pf
}
func (pf *EntityField) DefaultValue(val any) *EntityField { pf.Default = val; return pf }
func (pf *EntityField) DescText(desc string) *EntityField { pf.Desc = desc; return pf }
func (pf *EntityField) ExampleVal(val any) *EntityField   { pf.Example = val; return pf }

func Present(obj any, p *Entity, options ...H) H {
	out := H{}
	if obj == nil {
		return out
	}

	opts := H{}
	if len(options) > 0 {
		opts = options[0]
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, f := range p.Fields {
		if f.Condition != nil && !f.Condition(obj, opts) {
			continue
		}

		var val any
		if f.Func != nil {
			val = f.Func(obj)
		} else {
			fieldVal := v.FieldByName(f.Name)
			if !fieldVal.IsValid() || (fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil()) || fieldVal.IsZero() {
				val = f.Default
			} else {
				val = fieldVal.Interface()
			}
		}

		if f.Presenter != nil {
			val = serializeNested(val, f.Presenter, opts)
		}

		out[f.JSONKey] = val
	}
	return out
}

func PresentSlice(slice any, p *Entity, options ...H) []any {
	if slice == nil {
		return []any{}
	}

	v := reflect.ValueOf(slice)
	if v.Kind() != reflect.Slice {
		return []any{}
	}

	opts := H{}
	if len(options) > 0 {
		opts = options[0]
	}

	arr := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		item := v.Index(i).Interface()
		arr[i] = Present(item, p, opts)
	}
	return arr
}

func serializeNested(val any, presenter *Entity, options ...H) any {
	if val == nil {
		return nil
	}
	rv := reflect.ValueOf(val)

	opts := H{}
	if len(options) > 0 {
		opts = options[0]
	}

	switch rv.Kind() {
	case reflect.Slice:
		arr := []any{}
		for i := 0; i < rv.Len(); i++ {
			item := rv.Index(i).Interface()
			rt := reflect.ValueOf(item)
			if rt.Kind() == reflect.Ptr {
				arr = append(arr, Present(item, presenter, opts))
			} else {
				// Create a pointer to the item for Present
				itemVal := reflect.ValueOf(item)
				ptrVal := reflect.New(itemVal.Type())
				ptrVal.Elem().Set(itemVal)
				arr = append(arr, Present(ptrVal.Interface(), presenter, opts))
			}
		}
		return arr
	case reflect.Ptr, reflect.Struct:
		return Present(val, presenter, opts)
	case reflect.Map:
		m, ok := val.(map[string]any)
		if !ok {
			return nil
		}
		out := H(m)
		for k, v := range m {
			rv2 := reflect.ValueOf(v)
			if rv2.Kind() == reflect.Struct || rv2.Kind() == reflect.Ptr {
				out[k] = Present(v, presenter, opts)
			} else {
				out[k] = v
			}
		}
		return out
	default:
		return val
	}
}
