package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type AnimeName struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}

type DbsAnimeNames interface {
}

type PgDbsAnimeNames struct {
	db DbService
}

var dbsAnimeNamesInstance *PgDbsAnimeNames

func NewDbsAnimeNames(db DbService) DbsAnimeNames {
	if dbsAnimeNamesInstance != nil {
		return dbsAnimeNamesInstance
	}

	newDbsAnimeNames := &PgDbsAnimeNames{
		db: db,
	}
	dbsAnimeNamesInstance = newDbsAnimeNames

	return dbsAnimeNamesInstance
}

func InsertAnimeName(tx *sql.Tx, params *AnimeName) (*AnimeName, error) {
	result := &AnimeName{}

	query := `INSERT INTO anime_names (id, name)
	VALUES ($1, $2)
	RETURNING id, created_at, updated_at, name`

	err := tx.QueryRow(
		query,
		uuid.New(),
		params.Name,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Name,
	)
	if err != nil {
		log.Printf("error: DbsAnimeNames InsertAnimeName: Query: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectAnimeNameByName(tx *sql.Tx, params *AnimeName) (*AnimeName, error) {
	result := &AnimeName{}

	query := `SELECT * FROM anime_names
	WHERE name = $1`

	err := tx.QueryRow(query, params.Name).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Name,
	)
	if err != nil {
		log.Printf("error: DbsAnimeNames SelectAnimeNameByName: Query: %v", err)
		return nil, err
	}

	return result, nil
}
