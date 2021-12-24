package structenh

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

func Stringify(record interface{}) string {
	return asString(record, true)
}

func StringifyValue(record interface{}) string {
	return asString(record, false)
}

func asString(record interface{}, withAddress bool) string {
	t := reflect.TypeOf(record)
	v := reflect.ValueOf(record)
	s := ""
	//s += fmt.Sprintf("(%s)", t.Kind().String())
	switch t.Kind() {
	case reflect.Ptr:
		if withAddress {
			s += fmt.Sprintf("&0x%x", v.Pointer())
		}
		if t.Elem().Kind() != reflect.Struct {
			s += asString(v.Interface(), withAddress)
			return s
		}
		t = t.Elem()
		v = v.Elem()
		fallthrough
	case reflect.Struct:
		s += "{"
		if !v.IsValid() {
			return "(nil/zero)"
		}
		fieldNum := v.NumField()
		for i := 0; i < fieldNum; i++ {
			fieldType := t.Field(i)
			fieldValue := v.Field(i)
			s += fieldType.Name + ": "
			if !fieldValue.CanInterface() {
				s += "(can not read)"
			} else {
				s += asString(fieldValue.Interface(), withAddress)
			}
			if i != fieldNum-1 {
				s += ", "
			}
		}
		s += "}"
	case reflect.Map:
		if withAddress {
			s += fmt.Sprintf("map[0x%x ", v.Pointer())
		} else {
			s += "map[ "
		}

		lm := len(v.MapKeys())
		j := 0
		keyList := sMapKeys(v.MapKeys())
		sort.Sort(keyList)
		for _, key := range keyList {
			elemValue := v.MapIndex(key)
			s += key.String() + ": " + asString(elemValue.Interface(), withAddress)
			if lm != j+1 {
				s += ", "
			}
			j += 1
		}
		s += "]"
	case reflect.Array, reflect.Slice:
		if t.Kind() == reflect.Slice && withAddress {
			s += fmt.Sprintf("<0x%x ", v.Pointer())
		} else {
			s += "< "
		}

		la := v.Len()
		for j := 0; j < la; j++ {
			elemValue := v.Index(j)
			s += asString(elemValue.Interface(), withAddress)
			if la != j+1 {
				s += ", "
			}
		}
		s += " >"
	default:
		s += fmt.Sprint(record)
	}

	return s
}

type sMapKeys []reflect.Value

func (sbi sMapKeys) Len() int {
	return len(sbi)
}

func (sbi sMapKeys) Swap(i, j int) {
	sbi[i], sbi[j] = sbi[j], sbi[i]
}

func (sbi sMapKeys) Less(i, j int) bool {
	return strings.Compare(sbi[i].String(), sbi[j].String()) < 0
}
