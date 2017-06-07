package common

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

var typeRegistry = make(map[string]reflect.Type)
var typeInfos = make(map[string]*typeInfo)

type typeReference struct {
	field string
	kind  string
}

type typeInfo struct {
	kind       string
	references []*typeReference
	t          reflect.Type
	tptr       reflect.Type
}

func RegisterType(name string, dst interface{}) {
	st := reflect.TypeOf(dst)
	ost := st
	if st.Kind() == reflect.Ptr {
		st = st.Elem()
	} else {
		panic(fmt.Sprintf("expected pointer but got %T", dst))
	}
	typeRegistry[name] = st
	ti := typeInfo{
		kind:       name,
		references: make([]*typeReference, 0),
		t:          st,
		tptr:       ost,
	}
	for i := 0; i < st.NumField(); i++ {
		cur := st.Field(i)
		nn := cur.Tag.Get("meccano")
		if nn != "" {
			ti.references = append(ti.references,
				&typeReference{field: cur.Name, kind: nn})
		}
	}
	typeInfos[name] = &ti
}

func makeInstance(name string) interface{} {
	v := reflect.New(typeInfos[name].t).Elem()
	return v.Interface()
}

func makeSlice(name string, len int) reflect.Value {
	val := typeInfos[name].tptr
	return reflect.MakeSlice(reflect.SliceOf(val), len, len)
}

func MakePtrSlice(name string, len int) interface{} {
	val := typeInfos[name].tptr
	slice := reflect.MakeSlice(reflect.SliceOf(val), len, len)
	x := reflect.New(slice.Type())
	x.Elem().Set(slice)
	return x.Interface()
}

type key struct {
	Kind string
	Id   int64
}

type EntityContext struct {
	mapEntities map[key]interface{}
	kinds       map[string]bool
	visited     map[key]bool
	context     context.Context
}

func NewEntityContext(c context.Context) *EntityContext {
	ec := &EntityContext{context: c}
	ec.mapEntities = make(map[key]interface{})
	ec.kinds = make(map[string]bool)
	ec.visited = make(map[key]bool)
	return ec
}

func (ec *EntityContext) Add(dst interface{}) {
	s := reflect.ValueOf(dst)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	name := s.Type().Name()
	ec.kinds[name] = true
	f := s.FieldByName("Id")
	if !f.IsValid() {
		panic(fmt.Sprintf("field %q not found %+v %+v", "Id", dst, name))
	}
	ec.mapEntities[key{name, f.Int()}] = dst
}

func (ec *EntityContext) GetStringForField(dst interface{}, path string) (string, error) {
	pieces := strings.SplitN(path, ".", 2)
	val, err := ec.GetEntityForField(dst, pieces[0])
	if err != nil {
		return "", err
	}
	if val != nil {
		return ec.getStringValue(val, pieces[1])
	}
	return "", nil
}

func CapitalizeFirstLetter(str string) string {
	for i, l := range str {
		if unicode.IsLetter(l) {
			if i > 0 {
				return str[:i] + strings.ToUpper(str[i:i+1]) + str[i+1:]
			} else {
				return strings.ToUpper(str[:1]) + str[1:]
			}
		}
	}
	return ""
}

func (ec *EntityContext) GetEntityForField(dst interface{}, field string) (interface{}, error) {
	s := reflect.ValueOf(dst)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	name := s.Type().Name()
	field = CapitalizeFirstLetter(field)
	tr := ec.findReference(name, field)
	if tr == nil {
		return nil, fancyHandleError(fmt.Errorf("entity of type %s has no reference %s", name, field))
	}
	kind := tr.kind
	f := s.FieldByName(field)
	if !f.IsValid() {
		return nil, fancyHandleError(fmt.Errorf("field %q not found in %s", field, name))
	}
	if f.Int() == 0 {
		return nil, nil
	}
	k := key{kind, f.Int()}
	if val, exists := ec.mapEntities[k]; exists {
		return val, nil
	} else {
		return nil, fancyHandleError(fmt.Errorf("entity not found %+v", k))
	}
}

func (ec *EntityContext) Load() {
	keys := ec.findKeys()
	if len(keys) > 0 {
		ec.innerLoad(keys)
	}
	// for _, v := range ec.mapEntities {
	// 	fmt.Printf("k = %+T %+v\n", v, v)
	// }
}

func (ec *EntityContext) dbGet(curkeys []key, slice reflect.Value) error {
	toload := make([]*datastore.Key, len(curkeys))
	for i, k := range curkeys {
		toload[i] = datastore.NewKey(ec.context, k.Kind, "", k.Id, nil)
	}
	err := datastore.GetMulti(ec.context, toload, slice.Interface())
	if err != nil {
		if multiErr, ok := err.(appengine.MultiError); ok {
			for i := 0; i < len(multiErr); i++ {
				if multiErr[i] == nil {
					s := slice.Index(i).Interface()
					SetValueInt64(s, "Id", toload[i].IntID())
					ec.Add(s)
				}
			}
		} else {
			return err
		}
	} else {
		for i := 0; i < slice.Len(); i++ {
			s := slice.Index(i).Interface()
			SetValueInt64(s, "Id", toload[i].IntID())
			ec.Add(s)
		}
	}
	return nil
}

func (ec *EntityContext) innerLoad(keys []key) {
	groups := make(map[string][]key, 0)
	for _, k := range keys {
		if groups[k.Kind] == nil {
			groups[k.Kind] = make([]key, 0)
		}
		groups[k.Kind] = append(groups[k.Kind], k)
	}
	for kind, curkeys := range groups {
		slice := makeSlice(kind, len(curkeys))
		if err := ec.dbGet(curkeys, slice); err != nil {
			fmt.Printf("load err = %+v\n", err)
		}
	}
}

func (ec *EntityContext) findKeys() []key {
	keys := make(map[key]bool, 0)
	res := make([]key, 0)
	for k, v := range ec.mapEntities {
		if ti, ok := typeInfos[k.Kind]; ok {
			if len(ti.references) == 0 {
				continue
			}
			if _, ok := ec.visited[k]; ok {
				continue
			}
			ec.visited[k] = true
			s := reflect.ValueOf(v)
			if s.Kind() == reflect.Ptr {
				s = s.Elem()
			}
			for _, ref := range ti.references {
				f := s.FieldByName(ref.field)
				if f.IsValid() && f.Int() != 0 {
					k := key{ref.kind, f.Int()}
					if _, ok = ec.mapEntities[k]; !ok {
						keys[k] = true
						res = append(res, k)
					}
				}
			}
		}
	}
	return res
}

func (ec *EntityContext) getStringValue(dst interface{}, field string) (string, error) {
	s := reflect.ValueOf(dst)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}

	field = CapitalizeFirstLetter(field)
	f := s.FieldByName(field)
	if !f.IsValid() {
		return "", fancyHandleError(fmt.Errorf("field %+v not found in %+v", field, s.Type().Name()))
	}
	val := ""
	switch f.Kind() {
	case reflect.String:
		val = f.String()
	case reflect.Int:
		val = strconv.FormatInt(f.Int(), 10)
	case reflect.Int64:
		val = strconv.FormatInt(f.Int(), 10)
	default:
		return "", fancyHandleError(fmt.Errorf("type not supported %+v", f.Kind()))
	}
	return val, nil
}

func (ec *EntityContext) findReference(name string, field string) *typeReference {
	if ti, ok := typeInfos[name]; ok {
		for _, ref := range ti.references {
			if ref.field == field {
				return ref
			}
		}
	}
	return nil
}

func fancyHandleError(err error) error {
	return err
	//	pc, fn, line, _ := runtime.Caller(1)
	//
	//	return fmt.Errorf("%s[%s:%d] %v", runtime.FuncForPC(pc).Name(), fn, line, err)
}
