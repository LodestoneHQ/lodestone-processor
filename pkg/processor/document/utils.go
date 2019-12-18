package document

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"
)

func castToTime(val interface{}) time.Time {

	rt := reflect.TypeOf(val)
	if rt == nil {
		return time.Time{}
	}
	switch rt.Kind() {
	case reflect.Slice:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not a []string", val)
			return time.Time{}
		}

		t, err := time.Parse(time.RFC3339, valSlice[0])
		if err != nil {
			return time.Time{}
		}
		return t

	case reflect.Array:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not an array", val)
			return time.Time{}
		}
		t, err := time.Parse(time.RFC3339, valSlice[0])
		if err != nil {
			return time.Time{}
		}
		return t

	case reflect.String:
		valStr, ok := val.(string)
		if !ok {
			log.Printf("%v is not a string", val)
			return time.Time{}
		}

		t, err := time.Parse(time.RFC3339, valStr)
		if err != nil {
			return time.Time{}
		}
		return t

	default:
		return time.Time{}
	}

	return time.Time{}
}

func castToString(val interface{}) string {
	rt := reflect.TypeOf(val)
	if rt == nil {
		return ""
	}
	switch rt.Kind() {
	case reflect.Slice:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not a []string", val)
			return ""
		}
		return fmt.Sprintf(strings.Join(valSlice, ", "))
	case reflect.Array:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not an array", val)
			return ""
		}
		return fmt.Sprintf(strings.Join(valSlice, ", "))
	case reflect.String:
		valStr, ok := val.(string)
		if !ok {
			log.Printf("%v is not a string", val)
			return ""
		}
		return valStr
	default:
		return ""
	}
}

func castToStringArray(val interface{}) []string {
	rt := reflect.TypeOf(val)
	if rt == nil {
		return []string{}
	}

	switch rt.Kind() {
	case reflect.Slice:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not a []string", val)
			return []string{}
		}
		return valSlice
	case reflect.Array:
		valSlice, ok := val.([]string)
		if !ok {
			log.Printf("%v is not a []string", val)
			return []string{}
		}
		return valSlice
	case reflect.String:
		valStr, ok := val.(string)
		if !ok {
			log.Printf("%v is not a string", val)
			return []string{}
		}

		return []string{valStr}
	default:
		return nil
	}
}

//func stringHasValue(str string) bool {
//	return str != ""
//}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}
