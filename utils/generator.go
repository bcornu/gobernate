package utils

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// CORE OF THE DAO, FOR SAFETY REASONS, WILL PANIC ON ANY PROBLEM
func Generate(t reflect.Type) (creationQuery string, selectQuery string, insertQuery string, updateQuery string, deleteQuery string) {
	id := "Id"
	name := t.Name()
	sqlName := goToSql(name)
	sqlId := goToSql(id)
	fmt.Printf("generating requests for type: %s\n", name)

	//access value in case of pointer
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	//ignore non-struct values
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("not a struct, skipping %s\n", name))
	}

	hasId := false
	index := 1

	creationQuery = "CREATE TABLE " + sqlName + " ("
	selectQuery = "SELECT "
	insertQuery = "INSERT INTO " + sqlName + " ("
	updateQuery = "UPDATE " + sqlName + " SET "
	deleteQuery = "DELETE FROM " + sqlName + " WHERE " + sqlId + " = $1 "

	_insert1 := ""
	_insert2 := ""

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldName := f.Name
		kind := f.Type.Kind()
		isPtr := false
		if kind == reflect.Ptr {
			kind = f.Type.Elem().Kind()
			isPtr = true
		}
		if kind == reflect.Slice || kind == reflect.Array {
			panic(fmt.Sprintf("cannot deal with arrays yet %s:%s\n", name, fieldName))
		}

		sqlType := ""

		sqlFieldName := goToSql(fieldName)

		if kind == reflect.Struct && !KnownStructs(f.Type) {
			referenceName := ""
			if isPtr {
				referenceName = goToSql(f.Type.Elem().Name())
			} else {
				referenceName = goToSql(f.Type.Name())
			}
			sqlType = "int references " + referenceName + "(id)"
			sqlFieldName += "_id"
		} else {
			switch kind {
			case reflect.Bool:
				sqlType = "boolean"
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				sqlType = "integer"
			case reflect.Float32, reflect.Float64:
				sqlType = "float"
			case reflect.String:
				sqlType = "text"
			default:
				switch f.Type {
				case reflect.TypeOf(time.Time{}):
					sqlType = "timestamp without time zone"
				}
			}
		}

		if fieldName == id {
			hasId = true
			creationQuery += "\n" + sqlFieldName + " bigserial PRIMARY KEY,"
		} else {
			_insert1 += sqlFieldName + ", "
			_insert2 += "$" + fmt.Sprintf("%+v", index) + ", "
			creationQuery += "\n" + sqlFieldName + " " + sqlType + ","
			updateQuery += sqlFieldName + " = $" + fmt.Sprintf("%+v", index) + ", "
			index += 1
		}
		selectQuery += sqlFieldName + ", "
	}

	creationQuery = creationQuery[:len(creationQuery)-1] + ")"

	selectQuery = selectQuery[:len(selectQuery)-2]
	selectQuery += " FROM " + sqlName

	insertQuery += _insert1[:len(_insert1)-2] + ") VALUES (" + _insert2[:len(_insert2)-2] + ") RETURNING id"

	updateQuery = updateQuery[:len(updateQuery)-2]
	updateQuery += " WHERE " + sqlId + " = $" + fmt.Sprintf("%+v", index)

	if !hasId {
		panic("missing id from struct")
	}

	return creationQuery, selectQuery, insertQuery, updateQuery, deleteQuery
}

func goToSql(input string) string {
	res := ""
	re := regexp.MustCompile("^([A-Z]+)")
	res = re.ReplaceAllStringFunc(input, func(m string) string {
		return strings.ToLower(m)
	})
	re2 := regexp.MustCompile("([A-Z]+)")
	res = re2.ReplaceAllStringFunc(res, func(m string) string {
		return "_" + strings.ToLower(m)
	})
	return res
}

func KnownStructs(typ reflect.Type) bool {
	return typ == reflect.TypeOf(time.Time{})
}
