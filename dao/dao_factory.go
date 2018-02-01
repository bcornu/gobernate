package dao

import (
	"reflect"
)

var DAOs map[reflect.Type]GenericDAO = make(map[reflect.Type]GenericDAO)

func GetDAO(target interface{}) (GenericDAO, error) {
	typ := reflect.TypeOf(target)
	return getDAO(typ)
}

func getDAO(typ reflect.Type) (GenericDAO, error) {
	res, ok := DAOs[typ]
	if !ok {
		err := newDAO(typ)
		if err != nil {
			return GenericDAO{}, err
		}
	}
	res, ok = DAOs[typ]
	return res, nil
}

func newDAO(typ reflect.Type) error {
	res := GenericDAO{targetType: typ}
	err := res.init()
	if err != nil {
		return err
	}
	DAOs[res.targetType] = res
	return nil
}
