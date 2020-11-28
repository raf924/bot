package utils

import "reflect"

func AllValues(m map[string]interface{}) []interface{} {
	var values []interface{}
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func AllKeys(m interface{}) []string {
	v := reflect.ValueOf(m).MapKeys()
	var keys []string
	for _, k := range v {
		keys = append(keys, k.String())
	}
	return keys
}
