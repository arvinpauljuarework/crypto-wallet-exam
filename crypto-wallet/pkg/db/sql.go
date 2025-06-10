package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	once     sync.Once
	dbPool   *sql.DB
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
)

func DB() *sql.DB {
	once.Do(func() {
		var err error
		Host = "localhost"
		Port = "5432"
		User = "postgres"
		Password = "apple"
		DBName = "crypto-wallet"
		SSLMode = "disable"
		connectionConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", Host, Port, User, Password, DBName, SSLMode)
		dbPool, err = sql.Open("postgres", connectionConfig)
		if err != nil {
			fmt.Println(errors.Wrap(err, "error opening database connection")) // can put to logs or audit
		}

		dbPool.SetMaxOpenConns(5)
		dbPool.SetMaxIdleConns(5)
		dbPool.SetConnMaxLifetime(5 * time.Minute)
	})

	return dbPool
}
