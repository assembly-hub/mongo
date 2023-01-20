// package mongo
package mongo

import (
	"reflect"

	"github.com/assembly-hub/basics/set"
)

func Struct2Map(raw interface{}, excludeKey ...string) map[string]interface{} {
	dataValue := reflect.ValueOf(raw)
	if dataValue.Type().Kind() != reflect.Struct && dataValue.Type().Kind() != reflect.Ptr {
		panic("data type must be struct or struct ptr")
	}

	if dataValue.Type().Kind() == reflect.Ptr {
		dataValue = dataValue.Elem()
	}

	if dataValue.Type().Kind() != reflect.Struct {
		panic("data type must be struct or struct ptr")
	}

	s := set.Set[string]{}
	s.Add(excludeKey...)
	m := map[string]interface{}{}
	for i := 0; i < dataValue.NumField(); i++ {
		colName := dataValue.Type().Field(i).Tag.Get("bson")
		if colName == "" || !dataValue.Type().Field(i).IsExported() {
			continue
		}

		if colName == "_id" {
			if dataValue.Field(i).Kind() == reflect.String && dataValue.Field(i).String() == "" {
				continue
			}

			if dataValue.Field(i).Kind() == reflect.Array {
				var objID ObjectID
				b := dataValue.Field(i).Bytes()
				copy(objID[:], b)
				if objID.IsZero() {
					continue
				}
			}
		}

		if !s.Has(colName) {
			m[colName] = dataValue.Field(i).Interface()
		}
	}
	return m
}

func TransSession(cli *Client, fn func(sessionCtx SessionContext) error) error {
	return cli.NewSession(fn)
}
