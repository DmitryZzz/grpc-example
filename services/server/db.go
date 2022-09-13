package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	_ "github.com/lib/pq"
)

type DBManager struct {
	db  *sql.DB
	err error
}

func (m *DBManager) Connect(conf *PGConfig) {
	var connInfo string = fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=disable",
		conf.host, conf.port, conf.db, conf.username, conf.password)

	m.err = retry.Do(
		func() error {
			db, err := sql.Open("postgres", connInfo)
			if err != nil {
				return err
			}

			err = db.Ping()
			if err != nil {
				db.Close()
				return err
			}

			m.db = db
			return nil
		},
		retry.Attempts(5),
		retry.Delay(time.Second),
		retry.DelayType(retry.BackOffDelay),
	)
}

func (m *DBManager) Close() {
	m.db.Close()
}
