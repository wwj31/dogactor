//Copyright (c) 2017 Tap4Fun.Co.Ltd. All rights reserved.
package structenh

import (
	"fmt"
	"reflect"
)

func ValueEqual(src, des interface{}) bool {
	return valueEqual(src, des) && valueEqual(des, src)
}

func valueEqual(src, des interface{}) bool {
	srcType, desType := reflect.TypeOf(src), reflect.TypeOf(des)

	//type mis-match
	if srcType != desType || srcType.Name() != desType.Name() {
		fmt.Printf("type missmatch: %s - %s", srcType.Name(), desType.Name())
		return false
	}

	if srcType == nil && desType == nil {
		return true
	}

	srcValue, desValue := reflect.ValueOf(src), reflect.ValueOf(des)
	if !srcValue.IsValid() || !desValue.IsValid() {
		fmt.Printf("%s value invalid", srcType.Name())
		return false
	}

	switch srcType.Kind() {
	case reflect.Ptr:
		if srcValue.Pointer() == desValue.Pointer() {
			return true
		}
		if srcType.Elem().Kind() != reflect.Struct {
			if !valueCmp(srcValue, desValue) {
				return false
			}
		}
		srcType = srcType.Elem()
		srcValue = srcValue.Elem()
		desType = desType.Elem()
		desValue = desValue.Elem()
		fallthrough
	case reflect.Struct:
		fieldNum := srcValue.NumField()
		for i := 0; i < fieldNum; i++ {
			st := srcType.Field(i)
			sv := srcValue.Field(i)
			_, ok := desType.FieldByName(st.Name)
			if !ok {
				fmt.Printf("%s field not found in des-Type", st.Name)
				return false
			}
			dv := desValue.FieldByName(st.Name)
			if !valueCmp(sv, dv) {
				fmt.Printf("%s value not same", st.Name)
				return false
			}
		}
	case reflect.Array, reflect.Slice:
		sl := srcValue.Len()
		dl := desValue.Len()
		if sl != dl {
			return false
		}
		for j := 0; j < sl; j++ {
			sv := srcValue.Index(j)
			dv := desValue.Index(j)
			if !valueCmp(sv, dv) {
				fmt.Printf("%s array value not same", srcType.Name())
				return false
			}
		}
	case reflect.Map:
		sk := sMapKeys(srcValue.MapKeys())
		dk := sMapKeys(desValue.MapKeys())

		if !valueCmp(reflect.ValueOf(sk), reflect.ValueOf(dk)) {
			fmt.Printf("%s map keys not same", srcType.Name())
			return false
		}
		for _, key := range sk {
			sv := srcValue.MapIndex(key)
			dv := desValue.MapIndex(key)
			if !valueCmp(sv, dv) {
				return false
			}
		}
	default:
		//fmt.Printf("%v -- %v valueEuqal? %v", srcValue, desValue, srcValue.Interface() == desValue.Interface())
		return srcValue.Interface() == desValue.Interface()
	}

	return true
}

func valueCmp(sv, dv reflect.Value) bool {
	sc := sv.CanInterface()
	dc := dv.CanInterface()
	if sc == dc {
		if sc && !valueEqual(sv.Interface(), dv.Interface()) {
			fmt.Printf("%v -- %v valueEuqal fail", sv, dv)
			return false
		}
	} else {
		fmt.Printf("canInterface not same")
		return false
	}
	return true
}
