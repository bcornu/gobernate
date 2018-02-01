package utils

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
)

var (
	instance     *sql.DB
	once         sync.Once
	dbConfig     DBConfig
	dbConfigured = false
)

type DBConfig struct {
	Host     string
	Port     string
	Database string
	Login    string
	Password string
}

func InitDB(config DBConfig) {
	dbConfig = config
	dbConfigured = true
}

func GetSession() *sql.DB {
	if !dbConfigured {
		panic(errors.New("you must set db infos by using utils.InitDB first"))
	}
	once.Do(func() {

		connectStr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v", dbConfig.Login, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.Database)

		db, err := sql.Open("postgres", connectStr)

		if err != nil {
			panic(err)
		}

		instance = db
	})
	return instance
}
