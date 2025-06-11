package database

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type DbService interface {
	Conn() *sql.DB
	MigrateFS(migrationFS fs.FS, dir string) error
}

type dbService struct {
	db *sql.DB
}

var dbInstance *dbService

func NewDb() (DbService, error) {
	if dbInstance != nil {
		return dbInstance, nil
	}

	conn, err := Open()
	if err != nil {
		log.Panicf("error: database NewDb Open: %v:", err)
		return nil, err
	}

	newDbService := &dbService{
		db: conn,
	}
	dbInstance = newDbService

	return dbInstance, nil
}

func Open() (*sql.DB, error) {
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		"postgres",  // user
		"postgres",  // password
		"localhost", // host
		"5433",      // port
		"postgres",  // db
	)

	conn, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	err = conn.Ping()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *dbService) Conn() *sql.DB {
	return s.db
}

func (s *dbService) MigrateFS(migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()

	return Migrate(s.db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	if err := goose.Up(db, dir); err != nil {
		return fmt.Errorf("migrate: %v", err)
	}

	return nil
}
