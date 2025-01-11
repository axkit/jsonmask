// Package jsonmask provides functionality to mask JSON data based on field metadata.
package jsonmask

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// StructMaskRules holds metadata for a structure.
type StructMaskRules struct {
	Rules []Rule
}

// Rule holds metadata for a single field of a structure.
type Rule struct {
	// Path is a JSON path to the field.
	Path string

	// Action is a value of the mask tag.
	// It can be a name of a custom masking function or "-" to delete the field.
	Action     string
	sliceLevel int // 0 - no slice, 1 - slice, 2 - slice of slices, etc.
}

// DefaultStructFieldTag is a default tag name for struct fields.
const DefaultStructFieldTag = "mask"

// JsonMaskerImpl provides functionality to mask JSON data based on field metadata
// and custom masking functions.
type JsonMaskerImpl struct {
	tag   string // tag name for struct fields
	funcs map[string]func(string) []byte
}

// New creates a new instance of JsonMaskerImpl.
func New() *JsonMaskerImpl {
	return NewWithMaskTag(DefaultStructFieldTag)
}

// NewWithMaskTag creates a new instance of JsonMaskerImpl with a custom tag name.
func NewWithMaskTag(tag string) *JsonMaskerImpl {
	jm := JsonMaskerImpl{
		tag:   DefaultStructFieldTag,
		funcs: make(map[string]func(string) []byte),
	}

	jm.AddFunc("upper", Upper)
	jm.AddFunc("lower", Lower)
	jm.AddFunc("initialChar", InitialChar)
	jm.AddFunc("truncate", Truncate)
	jm.AddFunc("null", Null)
	jm.AddFunc("email", Email)
	jm.AddFunc("first4", PrefixFn(4, false))
	jm.AddFunc("zero", Zero)

	return &jm
}

// AddFunc adds a masking function associated with a name.
func (jm *JsonMaskerImpl) AddFunc(name string, f func(string) []byte) {
	jm.funcs[name] = f
}

// ParseStruct extracts metadata fields from the given structure based on the provided tag.
func (jm *JsonMaskerImpl) ParseStruct(src any) StructMaskRules {
	res := StructMaskRules{
		Rules: jm.extractStructRules(src, ""),
	}

	for i := range res.Rules {
		res.Rules[i].sliceLevel = strings.Count(res.Rules[i].Path, ".#")
	}

	return res
}

// joinPath joins parent and child attribute names using JSON path separator.
func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func (jm *JsonMaskerImpl) extractStructRules(str any, parentAttr string) []Rule {

	var rules []Rule

	s := reflect.ValueOf(str)

	// dereference pointer types
	for s.Kind() == reflect.Ptr {
		s = reflect.New(s.Type().Elem()).Elem()
	}

	if s.Kind() != reflect.Struct {
		return nil
	}

	t := s.Type()

	for i := 0; i < s.NumField(); i++ {
		sfv := s.Field(i)
		sft := t.Field(i)
		if !sft.IsExported() {
			continue
		}
		rules = append(rules, jm.extractStructFieldRules(sfv, sft, parentAttr)...)
	}

	return rules
}

func (jm *JsonMaskerImpl) extractStructFieldRules(
	val reflect.Value, // original field value
	sf reflect.StructField, // original field type
	parentAttr string,
) []Rule {

	var rules []Rule

	for val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val = reflect.New(val.Type().Elem()).Elem()
			continue
		}
		val = val.Elem()
	}

	kind := val.Kind()
	isSlice := kind == reflect.Slice || kind == reflect.Array
	if isSlice {
		if val.Len() > 0 {
			val = val.Index(0)
		} else {
			val = reflect.New(val.Type().Elem()).Elem()
		}
	}

	kind = val.Kind()
	jsonAttrName, jsonMaskTag := jm.parseFieldTag(sf)

	if jsonMaskTag == "-" {
		// quick return if tag holds "-".
		return []Rule{{Path: joinPath(parentAttr, jsonAttrName), Action: jsonMaskTag}}
	}

	if !(kind == reflect.Ptr || kind == reflect.Slice || kind == reflect.Array || kind == reflect.Struct || kind == reflect.Map) {
		// quick return if no mask tag and it's basic type.
		if jsonMaskTag == "" {
			return nil
		}
		return []Rule{{Path: joinPath(parentAttr, jsonAttrName), Action: jsonMaskTag}}
	}

	if isSlice {
		jsonAttrName = joinPath(parentAttr, jsonAttrName+".#")
	}

	switch val.Kind() {
	case reflect.Struct:
		rules = append(rules, jm.extractStructRules(val.Interface(), jsonAttrName)...)
	case reflect.Slice:
		for val.Kind() == reflect.Slice {
			val = reflect.New(val.Type().Elem()).Elem()
			jsonAttrName += ".#"
		}
		rules = append(rules, jm.extractStructRules(val.Interface(), jsonAttrName)...)
	default:
		rules = append(rules, Rule{Path: joinPath(parentAttr, jsonAttrName), Action: sf.Tag.Get(jm.tag)})
	}

	return rules
}

func (jm *JsonMaskerImpl) parseFieldTag(field reflect.StructField) (string, string) {
	jsonAttr := field.Tag.Get("json")
	if jsonAttr == "" || jsonAttr[0] == ',' { // if json is tag empty or looks like ",omitempty"
		jsonAttr = field.Name
	} else if idx := strings.IndexByte(jsonAttr, ','); idx >= 0 {
		jsonAttr = jsonAttr[:idx]
	}
	return jsonAttr, field.Tag.Get(jm.tag)
}

// Mask applies masking to JSON based on the given rules.
func (jm *JsonMaskerImpl) Mask(data []byte, smr StructMaskRules) ([]byte, error) {
	return jm.mask(data, smr.Rules)
}

func (jm *JsonMaskerImpl) mask(data []byte, rules []Rule) ([]byte, error) {
	var err error

	for _, rule := range rules {
		if rule.sliceLevel == 0 {
			data, err = jm.maskSimplePath(data, rule.Path, rule.Action)
		} else {
			idx := strings.Index(rule.Path, ".#")
			if idx < 0 {
				return nil, errors.New("invalid json array path")
			}
			data, err = jm.rangeOverArray(data, rule, rule.Path[:idx+2], rule.Path[idx+2:])
		}
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (jm *JsonMaskerImpl) maskSimplePath(data []byte, path, action string) ([]byte, error) {

	if action == "-" {
		return sjson.DeleteBytes(data, path)
	}

	maskFunc, exists := jm.funcs[action]
	if !exists {
		return data, nil
	}
	value := gjson.GetBytes(data, path)
	maskedValue := maskFunc(value.Raw)
	return sjson.SetRawBytes(data, path, maskedValue)
}

// items.#.#.currency
// items.#.balances.#.currency
// items.#.balances.#.#.amount

func (jm *JsonMaskerImpl) rangeOverArray(data []byte, rule Rule, arrPath, arrItemPath string) ([]byte, error) {
	var err error

	arr := gjson.GetBytes(data, arrPath)
	if !arr.Exists() {
		return data, errors.New("json array not found")
	}

	var subArrPath, subArrItemPath string
	subArrIdx := strings.Index(arrItemPath, ".#")
	if subArrIdx >= 0 {
		subArrPath = arrItemPath[:subArrIdx+2]
		subArrItemPath = arrItemPath[subArrIdx+2:]
	}

	// range over array
	for i := 0; i < int(arr.Int()); i++ {
		path := strings.ReplaceAll(arrPath, "#", strconv.Itoa(i))
		if rule.Action == "-" {
			data, err = sjson.DeleteBytes(data, path+arrItemPath)
			if err != nil {
				return nil, err
			}
			continue
		}

		// if array has no sub-array
		if subArrIdx < 0 {
			value := gjson.GetBytes(data, path+arrItemPath)
			maskFunc, exists := jm.funcs[rule.Action]
			if !exists {
				continue
			}

			maskedValue := maskFunc(value.Raw)
			data, err = sjson.SetRawBytes(data, path+arrItemPath, maskedValue)
		} else {
			// if array has sub-array
			data, err = jm.rangeOverArray(data, rule, path+subArrPath, subArrItemPath)
		}
		if err != nil {
			return nil, err
		}

	}
	return data, nil
}

// Error definitions
var (
	ErrInvalidInput = errors.New("input must be a struct")
)
