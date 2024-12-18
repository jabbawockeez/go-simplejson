package simplejson

import (
	"encoding/json"
	"errors"
	"fmt"

	"log"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/primitive"
	// util "github.com/jabbawockeez/go-utils"
)

// returns the current implementation version
func Version() string {
	return "1.1"
}

type Json struct {
	data interface{}
}

// NewJson returns a pointer to a new `Json` object
// after unmarshaling `body` bytes
func NewJson(body []byte) (*Json, error) {
	j := new(Json)
	err := j.UnmarshalJSON(body)
	if err != nil {
		return nil, err
	}
	return j, nil
}

// New returns a pointer to a new, empty `Json` object
func New() *Json {
	return &Json{
		data: make(map[string]interface{}),
	}
}

// Interface returns the underlying data
func (j *Json) Interface() interface{} {
	return j.data
}

// Encode returns its marshaled data as `[]byte`
func (j *Json) Encode() ([]byte, error) {
	return j.MarshalJSON()
}

// EncodePretty returns its marshaled data as `[]byte` with indentation
func (j *Json) EncodePretty() ([]byte, error) {
	return json.MarshalIndent(&j.data, "", "  ")
}

// Implements the json.Marshaler interface.
func (j *Json) MarshalJSON() ([]byte, error) {
	return json.Marshal(&j.data)
}

// Set modifies `Json` map by `key` and `value`
// Useful for changing single key/value in a `Json` object easily.
func (j *Json) Set(key string, val interface{}) {
	m, err := j.Map()
	if err != nil {
		return
	}
	m[key] = val
}

// SetPath modifies `Json`, recursively checking/creating map keys for the supplied path,
// and then finally writing in the value
func (j *Json) SetPath(branch []string, val interface{}) {
	if len(branch) == 0 {
		j.data = val
		return
	}

	// in order to insert our branch, we need map[string]interface{}
	if _, ok := (j.data).(map[string]interface{}); !ok {
		// have to replace with something suitable
		j.data = make(map[string]interface{})
	}
	curr := j.data.(map[string]interface{})

	for i := 0; i < len(branch)-1; i++ {
		b := branch[i]
		// key exists?
		if _, ok := curr[b]; !ok {
			n := make(map[string]interface{})
			curr[b] = n
			curr = n
			continue
		}

		// make sure the value is the right sort of thing
		if _, ok := curr[b].(map[string]interface{}); !ok {
			// have to replace with something suitable
			n := make(map[string]interface{})
			curr[b] = n
		}

		curr = curr[b].(map[string]interface{})
	}

	// add remaining k/v
	curr[branch[len(branch)-1]] = val
}

// func (j *Json) SetPath1(val interface{}, branch ...string) {
//     j.SetPath(branch, val)
// }

// an Enhanced set
func (j *Json) EnSet(args ...interface{}) {
    if len(args) < 2 {
        panic("EnSet accept at least two arguments: key and value!")
    }

    v := args[len(args) - 1]
    var val interface{} 
    typ := reflect.TypeOf(v)

    for typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }

    if typ.Name() == "Json" {
        v = v.(*Json).Interface()
        typ = reflect.TypeOf(v)
    }

    if typ.Kind() == reflect.Slice {
        value := reflect.ValueOf(v)
        ia := make([]interface{}, value.Len())

        for i := 0; i < value.Len(); i++ {
            ia[i] = value.Index(i).Interface()
        }
        val = &ia
    } else {
        val = v
    }

    branch := make([]string, len(args) - 1)
    for i := 0; i < len(args) - 1; i++ {
        branch[i] = args[i].(string)
    }

    if len(branch) == 1 {
        j.Set(branch[0], val)
    } else {
        j.SetPath(branch, val)
    }
}

func (j *Json) Length() (length int) {
    typ := reflect.TypeOf(j.Interface())

    if typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }

    switch typ.Kind() {
    case reflect.Slice:
        length = len(j.MustArray())
    case reflect.Map:
        length = len(j.MustMap())
    case reflect.String:
        length = len(j.MustString())
    default:
        panic("Can not get length of " + typ.String())
    }
    return
}

// Del modifies `Json` map by deleting `key` if it is present.
func (j *Json) Del(key string) {
	m, err := j.Map()
	if err != nil {
		return
	}
	delete(m, key)
}

func (j *Json) DelIndex(index int) (err error) {
	ap, e := j.ArrayPtr()
	if e != nil {
		err = e
        return
	}
	*ap = append((*ap)[:index], (*ap)[index + 1:]...)

    return
}

func (j *Json) Insert(index int, val interface{}) {
	ap, err := j.ArrayPtr()
	if err != nil {
		return
	}

    if index > len(*ap) {
        panic("Insert index out of range!")
    }

	*ap = append(*ap, 0)
    copy((*ap)[index + 1:], (*ap)[index:])
    (*ap)[index] = val
}

// Get returns a pointer to a new `Json` object
// for `key` in its `map` representation
//
// useful for chaining operations (to traverse a nested JSON):
// 
//	js.Get("top_level").Get("dict").Get("value").Int()
func (j *Json) Get(key string) *Json {
    m, err := j.Map()
    if err == nil {
        if val, ok := m[key]; ok {
            if reflect.TypeOf(val).Kind() == reflect.Slice {
                arr := val.([]interface{})
                m[key] = &arr
                return &Json{&arr}
            }
            return &Json{val}
        }
    }
    return &Json{nil}
}

// GetPath searches for the item as specified by the branch
// without the need to deep dive using Get()'s.
//
//	js.GetPath("top_level", "dict")
//func (j *Json) GetPath(branch ...string) *Json {
//	jin := j
//	for _, p := range branch {
//		jin = jin.Get(p)
//	}
//	return jin
//}
func (j *Json) GetPath(branch ...interface{}) *Json {
	jin := j
    var ok bool
	for _, p := range branch {
        switch p.(type) {
        case string:
            if jin, ok = jin.CheckGet(p.(string)); !ok {
                return &Json{nil}
            }
        case int:
            jin = jin.GetIndex(p.(int))
        }
	}

	return jin
}

// GetIndex returns a pointer to a new `Json` object
// for `index` in its `array` representation
//
// this is the analog to Get when accessing elements of
// a json array instead of a json object:
//    js.Get("top_level").Get("array").GetIndex(1).Get("key").Int()
//func (j *Json) GetIndex(index int) *Json {
//	a, err := j.Array()
//	if err == nil {
//		if len(a) > index {
//			return &Json{a[index]}
//		}
//	}
//	return &Json{nil}
//}

func (j *Json) GetIndex(index int) *Json {
	ap, err := j.ArrayPtr()
	if err == nil {
		if len(*ap) > index {
			return &Json{(*ap)[index]}
		}
	}
	return &Json{nil}
}

// CheckGet returns a pointer to a new `Json` object and
// a `bool` identifying success or failure
//
// useful for chained operations when success is important:
// 
//	if data, ok := js.Get("top_level").CheckGet("inner"); ok {
//	    log.Println(data)
//	}
func (j *Json) CheckGet(key string) (*Json, bool) {
	m, err := j.Map()
	if err == nil {
		if val, ok := m[key]; ok {
            if reflect.TypeOf(val).Kind() == reflect.Slice {
                arr := val.([]interface{})
                m[key] = &arr
                return &Json{&arr}, true
            }
            return &Json{val}, true
        }
	}
	return nil, false
}

// Map type asserts to `map`
func (j *Json) Map() (map[string]interface{}, error) {
    var ok bool
    var m map[string]interface{}
    // reflect.TypeOf(j.data)
    switch j.data.(type) {
    case primitive.M:
        m, ok = map[string]interface{}((j.data).(primitive.M))
    case map[string]interface{}:
        m, ok = (j.data).(map[string]interface{})
    }
	// if m, ok := (j.data).(map[string]interface{}); ok {
	// 	return m, nil
	// }
	if ok {
		return m, nil
	}
	return nil, errors.New("type assertion to map[string]interface{} failed")
}

// Array type asserts to an `array`
func (j *Json) Array() ([]interface{}, error) {
    ap, err := j.ArrayPtr()
    return *ap, err

    //if a, ok := (j.data).([]interface{}); ok {
    //    return a, nil
    //}
    //return nil, errors.New("type assertion to []interface{} failed")
}

func (j *Json) ArrayPtr() (p *[]interface{}, err error) {
    typ := reflect.TypeOf(j.data)

    if typ.Kind() == reflect.Ptr {
        var ok bool
        if p, ok = j.data.(*[]interface{}); !ok {
            err = errors.New("Not interface slice pointer!")
        }
    } else if typ.Kind() == reflect.Slice {
        arr := j.data.([]interface{})
        p = &arr
        j.data = p
    } else {
        err = errors.New("Not slice or slice pointer!")
    }

    return 
}

// Bool type asserts to `bool`
func (j *Json) Bool() (bool, error) {
	if s, ok := (j.data).(bool); ok {
		return s, nil
	}
	return false, errors.New("type assertion to bool failed")
}

// String type asserts to `string`
func (j *Json) String() (string, error) {
	if s, ok := (j.data).(string); ok {
		return s, nil
	}
	return "", errors.New("type assertion to string failed")
}

// Bytes type asserts to `[]byte`
func (j *Json) Bytes() ([]byte, error) {
	if s, ok := (j.data).(string); ok {
		return []byte(s), nil
	}
	return nil, errors.New("type assertion to []byte failed")
}

// StringArray type asserts to an `array` of `string`
func (j *Json) StringArray() ([]string, error) {
	arr, err := j.Array()
	if err != nil {
		return nil, err
	}
	retArr := make([]string, 0, len(arr))
	for _, a := range arr {
		if a == nil {
			retArr = append(retArr, "")
			continue
		}
		s, ok := a.(string)
		if !ok {
			return nil, errors.New("type assertion to []string failed")
		}
		retArr = append(retArr, s)
	}
	return retArr, nil
}

// MustArray guarantees the return of a `[]interface{}` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
// 
//	for i, v := range js.Get("results").MustArray() {
//		fmt.Println(i, v)
//	}
func (j *Json) MustArray(args ...[]interface{}) []interface{} {
	var def []interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustArray() received too many arguments %d", len(args))
	}

	a, err := j.Array()
	if err == nil {
		return a
	}

	return def
}

// MustMap guarantees the return of a `map[string]interface{}` (with optional default)
//
// useful when you want to interate over map values in a succinct manner:
// 
//	for k, v := range js.Get("dictionary").MustMap() {
//		fmt.Println(k, v)
//	}
func (j *Json) MustMap(args ...map[string]interface{}) map[string]interface{} {
	var def map[string]interface{}

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustMap() received too many arguments %d", len(args))
	}

	a, err := j.Map()
	if err == nil {
		return a
	}

	return def
}

// MustString guarantees the return of a `string` (with optional default)
//
// useful when you explicitly want a `string` in a single value return context:
//
//	myFunc(js.Get("param1").MustString(), js.Get("optional_param").MustString("my_default"))
func (j *Json) MustString(args ...string) string {
	var def string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustString() received too many arguments %d", len(args))
	}

	s, err := j.String()
	if err == nil {
		return s
	}

	return def
}

// MustStringArray guarantees the return of a `[]string` (with optional default)
//
// useful when you want to interate over array values in a succinct manner:
// 
//	for i, s := range js.Get("results").MustStringArray() {
//		fmt.Println(i, s)
//	}
func (j *Json) MustStringArray(args ...[]string) []string {
	var def []string

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustStringArray() received too many arguments %d", len(args))
	}

	a, err := j.StringArray()
	if err == nil {
		return a
	}

	return def
}

// MustInt guarantees the return of an `int` (with optional default)
//
// useful when you explicitly want an `int` in a single value return context:
//
//	myFunc(js.Get("param1").MustInt(), js.Get("optional_param").MustInt(5150))
func (j *Json) MustInt(args ...int) int {
	var def int

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt() received too many arguments %d", len(args))
	}

	i, err := j.Int()
	if err == nil {
		return i
	}

	return def
}

// MustFloat64 guarantees the return of a `float64` (with optional default)
//
// useful when you explicitly want a `float64` in a single value return context:
//     
//	myFunc(js.Get("param1").MustFloat64(), js.Get("optional_param").MustFloat64(5.150))
func (j *Json) MustFloat64(args ...float64) float64 {
	var def float64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustFloat64() received too many arguments %d", len(args))
	}

	f, err := j.Float64()
	if err == nil {
		return f
	}

	return def
}

// MustBool guarantees the return of a `bool` (with optional default)
//
// useful when you explicitly want a `bool` in a single value return context:
//
//	myFunc(js.Get("param1").MustBool(), js.Get("optional_param").MustBool(true))
func (j *Json) MustBool(args ...bool) bool {
	var def bool

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustBool() received too many arguments %d", len(args))
	}

	b, err := j.Bool()
	if err == nil {
		return b
	}

	return def
}

// MustInt64 guarantees the return of an `int64` (with optional default)
//
// useful when you explicitly want an `int64` in a single value return context:
//
//	myFunc(js.Get("param1").MustInt64(), js.Get("optional_param").MustInt64(5150))
func (j *Json) MustInt64(args ...int64) int64 {
	var def int64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustInt64() received too many arguments %d", len(args))
	}

	i, err := j.Int64()
	if err == nil {
		return i
	}

	return def
}

// MustUInt64 guarantees the return of an `uint64` (with optional default)
//
// useful when you explicitly want an `uint64` in a single value return context:
//     
//	myFunc(js.Get("param1").MustUint64(), js.Get("optional_param").MustUint64(5150))
func (j *Json) MustUint64(args ...uint64) uint64 {
	var def uint64

	switch len(args) {
	case 0:
	case 1:
		def = args[0]
	default:
		log.Panicf("MustUint64() received too many arguments %d", len(args))
	}

	i, err := j.Uint64()
	if err == nil {
		return i
	}

	return def
}

func (j *Json) Append(val interface{}) {
    ap, err := j.ArrayPtr()
    //util.P(*ap, err)
    if err != nil {
        log.Panic(err)
        return
    }

    typ := reflect.TypeOf(val)

    for typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }

    if typ.Name() == "Json" {
        val = val.(*Json).Interface()
    }

    //*(ap.(*interface{})) = append((*(ap.(*interface{}))).([]interface{}), val)
    *ap = append(*ap, val)
    j.data = ap
}

func (j *Json) Extend(val interface{}) {
    typ := reflect.TypeOf(val)

    var arr *Json

    if typ.Name() != "*Json" {
        arr = FromStruct(val)
    } else {
        arr = val.(*Json)
    }

    for _, item := range arr.Items() {
        j.Append(item)
    }
}

//func (j *Json) Append1(key string, val interface{}) {
//    var err error
//    var arr []interface{}
//    var data interface{}
//
//    typ := reflect.TypeOf(val)
//
//    if (typ.Kind() == reflect.Ptr) {
//        typ = typ.Elem()
//    }
//
//    switch typ.Kind() {
//    case reflect.Struct:
//        data = util.StructToMap(val)
//    default:
//        data = val
//    }
//
//    if key == "" {
//        arr, err = j.Array()
//        j.data = append(arr, data)
//    } else {
//        arr, err = j.Get(key).Array()
//        j.Set(key, append(arr, data))
//    }
//
//    if err != nil {
//        log.Panic(err)
//    }
//
//    return
//}

func (j *Json) Keys() (keys []string) {
    for k, _ := range j.MustMap() {
        keys = append(keys, k)
    }

    return 
}

func (j *Json) Items() (result map[interface{}]*Json) {
	result = map[interface{}]*Json{}

    typ := reflect.TypeOf(j.Interface())

    if typ.Kind() == reflect.Ptr {
        typ = typ.Elem()
    }

    switch typ.Kind() {
    case reflect.Slice:
		arr, _ := j.Array()
        for idx, item := range arr {
			result[idx] = &Json{item}
		}
    case reflect.Map:
        for key, item := range j.MustMap() {
			result[key] = &Json{item}
		}
    }

    return
}

func (j *Json) RenameKey(old, new string) {
    if v, ok := j.CheckGet(old); ok {
        j.EnSet(new, v)
        j.Del(old)
    }
}

func (j *Json) GetStringArray(path ...interface{}) (v []string) {
    switch len(path) {
    case 0:
        v = j.MustStringArray()
    default:
        v = j.GetPath(path...).MustStringArray()
    }

    return 
}

func (j *Json) GetString(path ...interface{}) (v string) {
    switch len(path) {
    case 0:
        v = j.MustString()
    default:
        v = j.GetPath(path...).MustString()
    }

    return 
}

func (j *Json) GetInt(path ...interface{}) (v int) {
    switch len(path) {
    case 0:
        v = j.MustInt()
    default:
        v = j.GetPath(path...).MustInt()
    }

    return 
}

func (j *Json) GetInt64(path ...interface{}) (v int64) {
    switch len(path) {
    case 0:
        v = j.MustInt64()
    default:
        v = j.GetPath(path...).MustInt64()
    }

    return 
}

func (j *Json) GetFloat64(path ...interface{}) (v float64) {
    switch len(path) {
    case 0:
        v = j.MustFloat64()
    default:
        v = j.GetPath(path...).MustFloat64()
    }

    return 
}

func (j *Json) GetMap(path ...interface{}) (v map[string]interface{}) {
    switch len(path) {
    case 0:
        v = j.MustMap()
    default:
        v = j.GetPath(path...).MustMap()
    }

    return 
}

func (j *Json) GetBool(path ...interface{}) (v bool) {
    switch len(path) {
    case 0:
        v = j.MustBool()
    default:
        v = j.GetPath(path...).MustBool()
    }

    return 
}

func (j *Json) GetArray(path ...interface{}) (v []interface{}) {
    switch len(path) {
    case 0:
        v = j.MustArray()
    default:
        v = j.GetPath(path...).MustArray()
    }

    return 
}

func (j *Json) Kind() reflect.Kind {
    return reflect.TypeOf(j.Interface()).Kind()
}

// ToStruct convert json object to struct
func (j *Json) ToStruct(v interface{}) error {
    byt, e := j.Encode()
    if e != nil {
        panic(e)
    }

    return json.Unmarshal(byt, v)
}

func (j *Json) ToString() string {
    byt, e := j.Encode()
    if e != nil {
        panic(e)
    }

    return string(byt)
}

func (j *Json) ToStringPretty() string {
    s, _ := j.EncodePretty()

    return string(s)
}

func (j *Json) P() {
    fmt.Println(j.ToStringPretty())
    // logs.Info(j.ToStringPretty())
}

func FromStruct(v interface{}) *Json {
    byt, e := json.Marshal(v)
    if e != nil {
        panic(e)
    }

    j, e := NewJson(byt)
    if e != nil {
        panic(e)
    }

    return j
}

// convert string or byte array to Json object
func FromString(i interface{}) *Json {
    var s []byte

    switch i.(type) {
        case string:
            s = []byte(i.(string))
        case []byte:
            s = i.([]byte)
    }

    j, e := NewJson(s)
    if e != nil {
        panic(e)
    }

    return j
}

func StructToString(v interface{}) string {
    return FromStruct(v).ToString()
}

func StringToStruct(s interface{}, v interface{}) error {
	var str string

	switch s.(type) {
    case []byte:
        str = string(s.([]byte))
    default:
        str = fmt.Sprintf("%v", s)
    }
    return FromString(str).ToStruct(v)
}

func (j *Json) Clone() (result *Json) {
    result = FromString(j.ToString())
    return
}
