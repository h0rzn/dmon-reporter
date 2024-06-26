package store

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/h0rzn/dmon-reporter/config"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

const defaultPath = "./store/data.db"
const (
	MODE_PERSIST = iota
	MODE_RELAY
)

type SqliteProvider struct {
	db     *sql.DB
	buffer *Buffer[Data]
	mode   int
}

func (p *SqliteProvider) Init(config *config.Config) error {
	log.SetLevel(logrus.DebugLevel)

	p.buffer = NewBuffer[Data](5)

	_, err := os.OpenFile(config.Cache.DbPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to create database file @ %s", config.Cache.DbPath))
	}

	log.Info("using db file:", config.Cache.DbPath)
	db, err := sql.Open("sqlite3", config.Cache.DbPath)
	if err != nil {
		return err
	}
	p.db = db

	return p.initScheme()
}

func (p *SqliteProvider) initScheme() error {
	createQuery := `
		CREATE TABLE IF NOT EXISTS data 
		(
			id INTEGER NOT NULL PRIMARY KEY,
			dataset JSON,
			time DATETIME NOT NULL
		);
	`
	_, err := p.db.Exec(createQuery)
	return err
}

func (p *SqliteProvider) SetMode(mode int) {
	p.mode = mode
}

func (p *SqliteProvider) Push(set Data) error {
	ready := p.buffer.Push(set)
	if ready {
		data := p.buffer.Drop()
		return p.Write(data...)
	}
	return nil
}

func (p *SqliteProvider) WriteSingle(data Data) error {
	jsonData, err := json.Marshal(data.Content())
	if err != nil || len(jsonData) == 0 {
		return err
	}
	insertQuery := `INSERT INTO data VALUES (?, json(?), DATETIME(?));`
	result, err := p.db.Exec(insertQuery, nil, string(jsonData), data.When())
	if err != nil {
		return err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (p *SqliteProvider) Write(sets ...Data) error {
	var query bytes.Buffer
	values := []interface{}{}
	query.WriteString("INSERT INTO data(dataset, time) VALUES ")
	for i, set := range sets {
		jsonData, err := json.Marshal(set.Content())
		if err != nil || len(jsonData) == 0 {
			return err
		}
		values = append(values, string(jsonData), set.When())
		query.WriteString("(?, ?)")
		if i < len(sets)-1 {
			query.WriteString(", ")
		}
	}

	result, err := p.db.Exec(query.String(), values...)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	log.Debug("writing to cache, rows affected: ", rowsAffected)
	return nil
}

func (p *SqliteProvider) Fetch() ([]Data, error) {
	query := "SELECT * FROM data"
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sets []Data
	for rows.Next() {
		var set Data
		var id int
		err := rows.Scan(&id, &set.data, &set.when)
		if err != nil {
			return sets, err
		}
		sets = append(sets, set)
	}

	return sets, nil
}

func (p *SqliteProvider) Drop() error {
	query := "DELETE FROM data"
	_, err := p.db.Exec(query)
	return err
}

func (p *SqliteProvider) Close() {
	p.db.Close()
}
