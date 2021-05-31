package structenh

import (
	"fmt"
	"reflect"
	"time"
)

//both function should call with input type of struct or pointer of struct
func DeepCopy(origin interface{}) interface{} {
	if origin == nil {
		return nil
	}
	original := reflect.ValueOf(origin)
	cpy := reflect.New(original.Type()).Elem()
	copyRecursive(original, cpy)
	return cpy.Interface()
}

var _timeType = reflect.TypeOf(time.Time{})

func copyRecursive(original, cpy reflect.Value) {
	// handle according to original's Kind
	if !cpy.CanSet() {
		return
	}

	switch original.Kind() {
	case reflect.Ptr:
		// get the actual value being pointed to.
		originalValue := original.Elem()
		// if  it isn't valid, return.
		if !originalValue.IsValid() {
			return
		}
		cpy.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, cpy.Elem())
	case reflect.Interface:
		// get the value for the interface, not the pointer.
		originalValue := original.Elem()
		if !originalValue.IsValid() {
			return
		}
		// get the value by calling Elem().
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		cpy.Set(copyValue)
	case reflect.Struct:
		oriType := original.Type()

		// Go through each field of the struct and copy it.
		if oriType.ConvertibleTo(_timeType) {
			t := original.Convert(_timeType)
			cpy.Set(t)
			return
		}
		for i := 0; i < original.NumField(); i++ {
			if cpy.Field(i).CanSet() &&
				(oriType.Field(i).Tag.Get("json") != "-" || oriType.Field(i).Tag.Get("cpy") != "") {
				copyRecursive(original.Field(i), cpy.Field(i))
			}
		}
	case reflect.Slice:
		// Make a new slice and copy each element.
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i))
		}
	case reflect.Map:
		if original.IsNil() {
			return
		}
		cpy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			cpy.SetMapIndex(key, copyValue)
		}
	// Sets the actual values from here on.
	default:
		cpy.Set(original)
	}
}

func InterfacePresentation(record interface{}) string {
	if reflect.TypeOf(record).Kind() != reflect.Ptr {
		return fmt.Sprint(record)
	}

	s := "{ "
	structType := reflect.TypeOf(record).Elem()
	structValue := reflect.ValueOf(record).Elem()
	if !structValue.IsValid() {
		return "(nil/zero)"
	}
	fieldNum := structValue.NumField()
	for i := 0; i < fieldNum; i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)
		s += fieldType.Name + ": "
		if !fieldValue.CanInterface() {
			s += "(can not read)"
			continue
		}
		switch fieldType.Type.Kind() {
		case reflect.Struct:
			s += InterfacePresentation(fieldValue.Addr().Interface())
		case reflect.Array, reflect.Slice:
			s += "< "
			la := fieldValue.Len()
			for j := 0; j < la; j++ {
				elemValue := fieldValue.Index(j)
				s += InterfacePresentation(elemValue.Interface())
				if la != j+1 {
					s += ", "
				}
			}
			s += " >"
		case reflect.Map:
			s += "map[ "
			lm := len(fieldValue.MapKeys())
			j := 0
			for _, key := range fieldValue.MapKeys() {
				elemValue := fieldValue.MapIndex(key)
				s += key.String() + ": " + InterfacePresentation(elemValue.Interface())
				if lm != j+1 {
					s += ", "
				}
				j += 1
			}
			s += " ]"
		case reflect.Ptr:
			s += InterfacePresentation(fieldValue.Interface())
		default:
			s += fmt.Sprint(fieldValue)
		}
		if i != fieldNum-1 {
			s += ", "
		}
	}
	return s + " }"
}
