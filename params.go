package grape

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type FieldType string

const (
	String    FieldType = "string"
	Integer   FieldType = "integer"
	Float     FieldType = "float"
	BigDecimal FieldType = "bigdecimal"
	Numeric   FieldType = "numeric"
	Date      FieldType = "date"
	DateTime  FieldType = "datetime"
	Time      FieldType = "time"
	Boolean   FieldType = "boolean"
	JSON      FieldType = "json"
	Slice     FieldType = "slice"
)

type Param struct {
	Name       string
	Type       FieldType
	Validate   string
	RequiredOn []string
	Schema     *Params
	SliceType  FieldType
}

type Params struct {
	Fields []Param
}

type FieldBuilder struct {
	param  Param
	parent *Params
}

func NewParams() *Params { return &Params{Fields: []Param{}} }

func (p *Params) Requires(name string) *FieldBuilder {
	fb := &FieldBuilder{param: Param{Name: name, RequiredOn: []string{}}, parent: p}
	p.Fields = append(p.Fields, fb.param)
	return fb
}

func (p *Params) Optional(name string) *FieldBuilder {
	fb := &FieldBuilder{param: Param{Name: name}, parent: p}
	p.Fields = append(p.Fields, fb.param)
	return fb
}

func (f *FieldBuilder) On(endpoints ...string) *FieldBuilder {
	f.param.RequiredOn = append(f.param.RequiredOn, endpoints...)
	f.updateParent()
	return f
}
func (f *FieldBuilder) String() *FieldBuilder    { f.param.Type = String; f.updateParent(); return f }
func (f *FieldBuilder) Integer() *FieldBuilder   { f.param.Type = Integer; f.updateParent(); return f }
func (f *FieldBuilder) Float() *FieldBuilder     { f.param.Type = Float; f.updateParent(); return f }
func (f *FieldBuilder) BigDecimal() *FieldBuilder { f.param.Type = BigDecimal; f.updateParent(); return f }
func (f *FieldBuilder) Numeric() *FieldBuilder   { f.param.Type = Numeric; f.updateParent(); return f }
func (f *FieldBuilder) Date() *FieldBuilder      { f.param.Type = Date; f.updateParent(); return f }
func (f *FieldBuilder) DateTime() *FieldBuilder  { f.param.Type = DateTime; f.updateParent(); return f }
func (f *FieldBuilder) Time() *FieldBuilder      { f.param.Type = Time; f.updateParent(); return f }
func (f *FieldBuilder) Boolean() *FieldBuilder   { f.param.Type = Boolean; f.updateParent(); return f }
func (f *FieldBuilder) JSON() *FieldBuilder      { f.param.Type = JSON; f.updateParent(); return f }
func (f *FieldBuilder) Slice() *FieldBuilder     { f.param.Type = Slice; f.updateParent(); return f }
func (f *FieldBuilder) Validate(tag string) *FieldBuilder {
	f.param.Validate = tag
	f.updateParent()
	return f
}
func (f *FieldBuilder) WithSchema(s *Params) *FieldBuilder {
	f.param.Schema = s
	f.updateParent()
	return f
}
func (f *FieldBuilder) SliceOf(t FieldType, s *Params) *FieldBuilder {
	f.param.Type = Slice
	f.param.SliceType = t
	f.param.Schema = s
	f.updateParent()
	return f
}

func (f *FieldBuilder) updateParent() {
	for i := range f.parent.Fields {
		if f.parent.Fields[i].Name == f.param.Name {
			f.parent.Fields[i] = f.param
			return
		}
	}
}

type Input map[string]interface{}

func (i Input) String(name string) string {
	if v, ok := i[name].(string); ok {
		return v
	}
	return ""
}
func (i Input) Integer(name string, def int) int {
	if v, ok := i[name].(int); ok {
		return v
	}
	return def
}
func (i Input) Float(name string, def float64) float64 {
	if v, ok := i[name].(float64); ok {
		return v
	}
	return def
}
func (i Input) Boolean(name string, def bool) bool {
	if v, ok := i[name].(bool); ok {
		return v
	}
	return def
}
func (i Input) BigDecimal(name string) string {
	if v, ok := i[name].(string); ok {
		return v
	}
	return ""
}
func (i Input) Numeric(name string, def float64) float64 {
	if v, ok := i[name].(float64); ok {
		return v
	}
	if s, ok := i[name].(string); ok {
		// Try to parse string as float
		if f, err := parseFloat(s); err == nil {
			return f
		}
	}
	return def
}
func (i Input) Date(name string) string {
	if v, ok := i[name].(string); ok {
		return v
	}
	return ""
}
func (i Input) DateTime(name string) string {
	if v, ok := i[name].(string); ok {
		return v
	}
	return ""
}
func (i Input) Time(name string) string {
	if v, ok := i[name].(string); ok {
		return v
	}
	return ""
}
func (i Input) JSON(name string) interface{} {
	return i[name]
}

// Helper function to parse float from string
func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// ToModel maps Input data to a struct pointer, converting snake_case keys to PascalCase fields.
// It only sets non-nil values to avoid overwriting existing data.
func (i Input) ToModel(dst interface{}) {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		panic("dst must be non-nil pointer")
	}
	v = v.Elem()

	for k, val := range i {
		field := v.FieldByNameFunc(func(n string) bool {
			return stringsEqualFold(n, k)
		})
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		if val == nil {
			continue // don't overwrite with nil
		}

		fv := reflect.ValueOf(val)
		if fv.Type().AssignableTo(field.Type()) {
			field.Set(fv)
		} else if fv.Type().ConvertibleTo(field.Type()) {
			field.Set(fv.Convert(field.Type()))
		}
	}
}

// stringsEqualFold performs case-insensitive comparison of struct field name and JSON key.
// For example: "first_name" <-> "FirstName"
func stringsEqualFold(structField, jsonKey string) bool {
	s := toSnakeCase(structField)
	return s == jsonKey
}

// toSnakeCase converts PascalCase to snake_case
func toSnakeCase(str string) string {
	if str == "" {
		return ""
	}

	out := make([]rune, 0, len(str)*2)
	for i, r := range str {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && str[i-1] >= 'a' && str[i-1] <= 'z' {
				out = append(out, '_')
			}
			out = append(out, r-'A'+'a')
		} else {
			out = append(out, r)
		}
	}
	return string(out)
}

func (p *Params) BindAndValidate(raw map[string]interface{}, mode string) (Input, error) {
	out := Input{}

	for _, f := range p.Fields {
		val, ok := raw[f.Name]

		isRequired := false
		for _, r := range f.RequiredOn {
			if strings.TrimSpace(r) == mode {
				isRequired = true
				break
			}
		}

		if !ok {
			if isRequired {
				return nil, fmt.Errorf("missing required field '%s' for %s", f.Name, mode)
			}
			continue
		}

		switch f.Type {
		case String:
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be string", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(s, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = s
		case Integer:
			switch vv := val.(type) {
			case float64:
				i := int(vv)
				if f.Validate != "" {
					if err := validate.Var(i, f.Validate); err != nil {
						return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
					}
				}
				out[f.Name] = i
			case int:
				out[f.Name] = vv
			default:
				return nil, fmt.Errorf("field '%s' must be integer", f.Name)
			}
		case Float:
			fv, ok := val.(float64)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be float", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(fv, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = fv
		case BigDecimal:
			// BigDecimal can be a string representation of a decimal number
			switch vv := val.(type) {
			case string:
				if f.Validate != "" {
					if err := validate.Var(vv, f.Validate); err != nil {
						return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
					}
				}
				out[f.Name] = vv
			case float64:
				s := fmt.Sprintf("%.10f", vv)
				out[f.Name] = s
			default:
				return nil, fmt.Errorf("field '%s' must be bigdecimal (string or float)", f.Name)
			}
		case Numeric:
			// Numeric is similar to Float but accepts both float and string
			switch vv := val.(type) {
			case float64:
				if f.Validate != "" {
					if err := validate.Var(vv, f.Validate); err != nil {
						return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
					}
				}
				out[f.Name] = vv
			case string:
				if f.Validate != "" {
					if err := validate.Var(vv, f.Validate); err != nil {
						return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
					}
				}
				out[f.Name] = vv
			default:
				return nil, fmt.Errorf("field '%s' must be numeric (float or string)", f.Name)
			}
		case Date:
			// Date expects a string in date format
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be date string", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(s, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = s
		case DateTime:
			// DateTime expects a string in datetime format
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be datetime string", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(s, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = s
		case Time:
			// Time expects a string in time format
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be time string", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(s, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = s
		case Boolean:
			bv, ok := val.(bool)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be boolean", f.Name)
			}
			out[f.Name] = bv
		case JSON:
			// JSON can be a map, slice, or string containing JSON
			switch vv := val.(type) {
			case map[string]interface{}:
				if f.Schema != nil {
					nested, err := f.Schema.validateJSON(vv, mode)
					if err != nil {
						return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
					}
					out[f.Name] = nested
				} else {
					out[f.Name] = vv
				}
			case []interface{}:
				out[f.Name] = vv
			case string:
				// Try to parse as JSON string
				var parsed interface{}
				if err := json.Unmarshal([]byte(vv), &parsed); err != nil {
					return nil, fmt.Errorf("field '%s' must be valid JSON", f.Name)
				}
				out[f.Name] = parsed
			default:
				return nil, fmt.Errorf("field '%s' must be json (object, array, or json string)", f.Name)
			}
		case Slice:
			svals, ok := val.([]interface{})
			if !ok {
				return nil, fmt.Errorf("field '%s' must be array", f.Name)
			}
			if f.SliceType == JSON && f.Schema != nil {
				arr := make([]interface{}, 0, len(svals))
				for _, elem := range svals {
					m, ok := elem.(map[string]interface{})
					if !ok {
						return nil, fmt.Errorf("element in '%s' must be object", f.Name)
					}
					nested, err := f.Schema.validateJSON(m, mode)
					if err != nil {
						return nil, fmt.Errorf("element in '%s' validation failed: %w", f.Name, err)
					}
					arr = append(arr, nested)
				}
				out[f.Name] = arr
			} else {
				out[f.Name] = svals
			}
		default:
			out[f.Name] = val
		}
	}

	for k, v := range raw {
		if _, ok := out[k]; !ok {
			out[k] = v
		}
	}

	return out, nil
}

// BindAndValidateReader binds JSON from an io.Reader and validates
func (p *Params) BindAndValidateReader(reader io.Reader, mode string) (Input, error) {
	var raw map[string]interface{}
	if err := json.NewDecoder(reader).Decode(&raw); err != nil {
		return nil, err
	}
	return p.BindAndValidate(raw, mode)
}

// Convenience functions for popular frameworks
//
// Usage examples:
//
// For Gin:
//   input, err := schema.BindAndValidateReader(c.Request.Body, "create")
//
// For Echo:
//   input, err := schema.BindAndValidateReader(c.Request().Body, "create")
//
// For Chi:
//   input, err := schema.BindAndValidateReader(r.Body, "create")
//
// For Fiber:
//   input, err := schema.BindAndValidateReader(c.Body(), "create")
//
// For direct map usage:
//   input, err := schema.BindAndValidate(myMap, "create")

func (p *Params) validateJSON(raw map[string]interface{}, mode string) (map[string]interface{}, error) {
	b, _ := json.Marshal(raw)
	var parsed map[string]interface{}
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	for _, f := range p.Fields {
		val, ok := parsed[f.Name]

		isRequired := false
		for _, r := range f.RequiredOn {
			if strings.TrimSpace(r) == mode {
				isRequired = true
				break
			}
		}
		if !ok {
			if isRequired {
				return nil, fmt.Errorf("missing required field '%s' for %s", f.Name, mode)
			}
			continue
		}

		switch f.Type {
		case String:
			s, ok := val.(string)
			if !ok {
				return nil, fmt.Errorf("field '%s' must be string", f.Name)
			}
			if f.Validate != "" {
				if err := validate.Var(s, f.Validate); err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
			}
			out[f.Name] = s
		case Integer:
			switch vv := val.(type) {
			case float64:
				out[f.Name] = int(vv)
			case int:
				out[f.Name] = vv
			default:
				return nil, fmt.Errorf("field '%s' must be integer", f.Name)
			}
		case JSON:
			if f.Schema != nil {
				nested, err := f.Schema.validateJSON(val.(map[string]interface{}), mode)
				if err != nil {
					return nil, fmt.Errorf("field '%s' validation failed: %w", f.Name, err)
				}
				out[f.Name] = nested
			} else {
				out[f.Name] = val
			}
		default:
			out[f.Name] = val
		}
	}

	for k, v := range parsed {
		if _, ok := out[k]; !ok {
			out[k] = v
		}
	}
	return out, nil
}
