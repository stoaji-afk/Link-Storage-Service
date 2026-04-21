package db

import (
    "database/sql"
    _ "github.com/lib/pq" // PostgreSQL driver
)

type DB struct {
    *sql.DB
}

func New(url string) (*DB, error) {
    db, err := sql.Open("postgres", url)
    if err != nil {
        return nil, err
    }

    if err = db.Ping(); err != nil {
        return nil, err
    }

    return &DB{db}, nil
}

// Инициализация схемы БД
func (d *DB) Init() error {
    query := `
    CREATE TABLE IF NOT EXISTS links (
        id SERIAL PRIMARY KEY,
        short_code VARCHAR(255) UNIQUE NOT NULL,
        original_url TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        visits INTEGER DEFAULT 0
    )`

    _, err := d.Exec(query)
    return err
}