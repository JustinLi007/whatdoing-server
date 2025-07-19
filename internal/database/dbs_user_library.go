package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type UserLibrary struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uuid.UUID `json:"-"`
}

type DbsUserLibrary interface {
	CreateUserLibrary(reqUser *User) error              // TODO: needed?
	GetUserLibrary(reqUser *User) (*UserLibrary, error) // TODO: needed?
}

type PgDbsUserLibrary struct {
	db DbService
}

var dbsUserLibraryInstance *PgDbsUserLibrary

func NewDbsUserLibrary(db DbService) DbsUserLibrary {
	if dbsUserLibraryInstance != nil {
		return dbsUserLibraryInstance
	}

	newDbsUserLibrary := &PgDbsUserLibrary{
		db: db,
	}
	dbsUserLibraryInstance = newDbsUserLibrary

	return dbsUserLibraryInstance
}

func (d *PgDbsUserLibrary) CreateUserLibrary(reqUser *User) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsUserLibrary CreateList: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsUserLibrary CreateList: Rollback: %v", err)
		}
	}()

	userLib, err := SelectUserLibrary(tx, reqUser)
	if err == nil && userLib != nil {
		msg := "error: DbsUserLibrary already exist for user"
		log.Printf("%s", msg)
		return errors.New(msg)
	}

	err = InsertUserLibrary(tx, reqUser)
	if err != nil {
		log.Printf("error: DbsUserLibrary CreateList: InsertUserLibrary: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsUserLibrary CreateList: Commit: %v", err)
		return err
	}

	return nil
}

func (d *PgDbsUserLibrary) GetUserLibrary(reqUser *User) (*UserLibrary, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsUserLibrary GetUserLibrary: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsUserLibrary GetUserLibrary: Rollback: %v", err)
		}
	}()

	dbUserLibrary, err := SelectUserLibrary(tx, reqUser)
	if err != nil {
		log.Printf("error: DbsUserLibrary GetUserLibrary: SelectUserLibrary: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsUserLibrary GetUserLibrary: Commit: %v", err)
		return nil, err
	}

	return dbUserLibrary, nil
}

func InsertUserLibrary(tx *sql.Tx, reqUser *User) error {
	query := `INSERT INTO user_library (id, user_id)
	VALUES ($1, $2)`

	queryResult, err := tx.Exec(
		query,
		uuid.New(),
		reqUser.Id,
	)
	if err != nil {
		log.Printf("error: DbsUserLibrary InsertUserLibrary: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err == nil {
		if n == 0 {
			log.Printf("error: DbsUserLibrary InsertUserLibrary: RowsAffected: 0")
			return sql.ErrNoRows
		}
	}

	return nil
}

func SelectUserLibrary(tx *sql.Tx, reqUser *User) (*UserLibrary, error) {
	result := &UserLibrary{}

	query := `SELECT * FROM user_library WHERE user_id = $1`

	err := tx.QueryRow(
		query,
		reqUser.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.UserId,
	)
	if err != nil {
		log.Printf("error: DbsUserLibrary SelectUserLibraryByUserId: Scan: %v", err)
		return nil, err
	}

	return result, nil
}
