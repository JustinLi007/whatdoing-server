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
		log.Printf("error: Dbs: AnimeNames: InsertAnimeName: Query: %v", err)
		return nil, err
	}

	return result, nil
}

func InsertAnimeNameIfNotExist(tx *sql.Tx, reqAnimeName *AnimeName) (*AnimeName, error) {
	result := &AnimeName{}

	query := `
	WITH source(id, name) AS (
		VALUES($1::uuid, $2)
	), upsert AS (
		MERGE INTO anime_names AS an
		USING source AS src
			ON LOWER(an.name) = LOWER(src.name)
		WHEN MATCHED THEN
			UPDATE SET id = an.id
		WHEN NOT MATCHED THEN
			INSERT(id, name)
			VALUES(src.id, src.name)
		RETURNING an.id, an.created_at, an.updated_at, an.name
	)
	SELECT n.id, n.created_at, n.updated_at, n.name
	FROM upsert n
	JOIN rel_anime_anime_names alt ON n.id = alt.anime_names_id
	`

	if err := tx.QueryRow(
		query,
		uuid.New(),
		reqAnimeName.Name,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Name,
	); err != nil {
		log.Printf("error: Dbs: AnimeNames: InsertAnimeNameIfNotExist: Scan: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectAnimeNameByName(tx *sql.Tx, params *AnimeName) (*AnimeName, error) {
	result := &AnimeName{}

	query := `SELECT * FROM anime_names
	WHERE LOWER(name) = LOWER($1)`

	err := tx.QueryRow(query, params.Name).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Name,
	)
	if err != nil {
		log.Printf("error: Dbs: AnimeNames: SelectAnimeNameByName: Query: %v", err)
		return nil, err
	}

	return result, nil
}
