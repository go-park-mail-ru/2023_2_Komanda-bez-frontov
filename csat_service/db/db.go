package db

import (
 "database/sql"
 
_"github.com/jackc/pgx/stdlib"
)

// Database представляет подключение к базе данных
type Database struct {
 Conn *sql.DB
}

// NewDB открывает соединение с базой данных
func NewDB() (*Database, error) {
 connStr := "user=nofronts password=nofronts dbname=nofronts_csat host=localhost port=5432 sslmode=disable"
 db, err := sql.Open("pgx", connStr)
 if err != nil {
  return nil, err
 }
 if err = db.Ping(); err != nil {
  return nil, err
 }
 return &Database{Conn: db}, nil
}

// Close закрывает соединение с базой данных
func (db *Database) Close() {
 db.Conn.Close()
}