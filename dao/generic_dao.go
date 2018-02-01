package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bcornu/gobernate/utils"
	"reflect"
	"time"
)

type GenericDAO struct {
	targetType    reflect.Type
	creationQuery string
	selectQuery   string
	insertQuery   string
	updateQuery   string
	deleteQuery   string
}

func (m *GenericDAO) init() error {
	m.creationQuery, m.selectQuery, m.insertQuery, m.updateQuery, m.deleteQuery = utils.Generate(m.targetType)
	return nil
}

func (m *GenericDAO) Log() {
	fmt.Println("targetType")
	fmt.Println(m.targetType)
	fmt.Println("creationQuery")
	fmt.Println(m.creationQuery)
	fmt.Println("selectQuery")
	fmt.Println(m.selectQuery)
	fmt.Println("insertQuery")
	fmt.Println(m.insertQuery)
	fmt.Println("updateQuery")
	fmt.Println(m.updateQuery)
	fmt.Println("deleteQuery")
	fmt.Println(m.deleteQuery)
}

func (m *GenericDAO) assertOK() {
	if m.creationQuery == "" || m.selectQuery == "" || m.insertQuery == "" || m.updateQuery == "" || m.deleteQuery == "" {
		panic("missiong init, this should never happens")
	}
}

func (m *GenericDAO) GetOne(id int) (interface{}, error) {
	m.assertOK()
	db := utils.GetSession()
	query := m.selectQuery + " WHERE id = $1 "
	rows, err := db.Query(query, id)
	if err != nil {
		fmt.Println(query)
		return nil, err
	}
	for rows.Next() {
		mission, err := m.scan(rows)
		if err != nil {
			return nil, err
		}
		return mission, nil
	}
	return nil, errors.New("no item found")
}

func (m *GenericDAO) GetAll() (interface{}, error) {
	m.assertOK()
	db := utils.GetSession()

	res := reflect.MakeSlice(reflect.SliceOf(m.targetType), 0, 10)

	rows, err := db.Query(m.selectQuery)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		mission, err := m.scan(rows)
		if err != nil {
			return nil, err
		}
		res = reflect.Append(res, reflect.ValueOf(mission))
	}
	return res.Interface(), nil
}

func (m *GenericDAO) scan(rows *sql.Rows) (interface{}, error) {
	colsNum := m.targetType.NumField()
	res := reflect.New(m.targetType).Elem()
	vals := make([]interface{}, colsNum)
	keys := make([]interface{}, colsNum)
	for i := 0; i < colsNum; i++ {
		keys[i] = &vals[i]
	}
	err := rows.Scan(keys...)
	if err != nil {
		return nil, err
	}
	for i := 0; i < colsNum; i++ {
		f := res.Field(i)
		val := vals[i]
		fieldType := f.Type()
		isPtr := false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = f.Type().Elem()
			isPtr = true
		}
		switch f.Type().Kind() {
		case reflect.Bool:
			f.SetBool(val.(bool))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			f.SetInt(val.(int64))
		case reflect.Float32, reflect.Float64:
			f.SetFloat(val.(float64))
		case reflect.String:
			f.SetString(val.(string))
		default:
			switch fieldType {
			case reflect.TypeOf(time.Time{}):
				f.Set(reflect.ValueOf(val.(time.Time)))
			default:
				if val == nil {

				} else if reflect.TypeOf(val).Kind() == reflect.Int64 {
					tmpDAO, err := getDAO(fieldType)
					if err != nil {
						return nil, err
					}
					tmp, err := tmpDAO.GetOne(int(val.(int64)))
					if err != nil {
						return nil, err
					}
					if isPtr {
						reflectTmp := reflect.ValueOf(tmp)
						vp := reflect.New(reflectTmp.Type())
						vp.Elem().Set(reflectTmp)
						f.Set(vp)
					} else {
						f.Set(reflect.ValueOf(tmp))
					}
				} else {
					return nil, errors.New("no maaping found")
				}
			}
		}
	}
	return res.Interface(), nil
}

func (m *GenericDAO) Create() error {
	m.assertOK()
	db := utils.GetSession()
	_, err := db.Exec(m.creationQuery)
	if err != nil {
		return err
	}
	return nil
}

func (m *GenericDAO) Insert(value interface{}) (res interface{}, err error) {
	m.assertOK()
	db := utils.GetSession()
	colsNum := m.targetType.NumField()
	vals := make([]interface{}, colsNum-1)
	currentId := 0
	for i := 0; i < colsNum; i++ {
		f := m.targetType.Field(i)
		fieldType := f.Type
		isPtr := false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = f.Type.Elem()
			isPtr = true
		}
		if f.Name != "Id" {
			var val interface{}
			if isPtr {
				if reflect.ValueOf(value).Field(i).IsNil() {
					vals[currentId] = nil
					currentId += 1
					continue
				} else {
					val = reflect.ValueOf(value).Field(i).Elem().Interface()
				}
			} else {
				val = reflect.ValueOf(value).Field(i).Interface()
			}
			if fieldType.Kind() == reflect.Struct && !utils.KnownStructs(fieldType) {
				vals[currentId], err = createOrUpdate(val)
				if err != nil {
					return nil, err
				}
			} else {
				vals[currentId] = val
			}
			currentId += 1
		}
	}
	rows, err := db.Query(m.insertQuery, vals...)
	if err != nil {
		return nil, err
	}
	id := 0
	for rows.Next() {
		err = rows.Scan(&id)
	}
	res, err = m.GetOne(id)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *GenericDAO) Update(value interface{}) (res interface{}, err error) {
	m.assertOK()
	db := utils.GetSession()
	colsNum := m.targetType.NumField()
	vals := make([]interface{}, colsNum)
	currentId := 0
	objId := 0
	for i := 0; i < colsNum; i++ {
		f := m.targetType.Field(i)
		fieldType := f.Type
		isPtr := false
		if fieldType.Kind() == reflect.Ptr {
			fieldType = f.Type.Elem()
			isPtr = true
		}
		var val interface{}
		if isPtr {
			val = reflect.ValueOf(value).Field(i).Elem().Interface()
		} else {
			val = reflect.ValueOf(value).Field(i).Interface()
		}
		if f.Name == "Id" {
			vals[colsNum-1] = val
			objId = int(val.(int64))
		} else {
			if fieldType.Kind() == reflect.Struct && !utils.KnownStructs(fieldType) {
				vals[currentId], err = createOrUpdate(val)
				if err != nil {
					return nil, err
				}
			} else {
				vals[currentId] = val
			}
			currentId += 1
		}
	}
	_, err = db.Query(m.updateQuery, vals...)
	if err != nil {
		return nil, err
	}
	res, err = m.GetOne(objId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *GenericDAO) Delete(id int) error {
	m.assertOK()
	db := utils.GetSession()
	_, err := db.Exec(m.deleteQuery, id)
	return err
}

func createOrUpdate(obj interface{}) (int64, error) {
	childId := getStructId(obj)
	tmpDAO, err := getDAO(reflect.TypeOf(obj))
	if err != nil {
		return 0, err
	}
	if childId > 0 {
		tmpDAO.Update(obj)
		return childId, nil
	} else {
		tmp, err := tmpDAO.Insert(obj)
		if err != nil {
			return 0, err
		}
		return getStructId(tmp), nil
	}
}

func getStructId(obj interface{}) int64 {
	return reflect.ValueOf(obj).FieldByName("Id").Interface().(int64)
}
