package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type AnimeAltName struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AnimeId   uuid.UUID `json:"anime_id"`
	AnimeName AnimeName `json:"anime_name"`
}

type DbsAnimeAltNames interface {
	AddAltName(reqAltName *AnimeAltName) error
	DeleteAltNames(reqAltNames []*AnimeAltName) error
}

type PgDbsAnimeAltNames struct {
	db DbService
}

var dbsAnimeAltNamesInstance *PgDbsAnimeAltNames

func NewDbsAnimeAltNames(db DbService) DbsAnimeAltNames {
	if dbsAnimeAltNamesInstance != nil {
		return dbsAnimeAltNamesInstance
	}

	newDbsAnimeAltNames := &PgDbsAnimeAltNames{
		db: db,
	}
	dbsAnimeAltNamesInstance = newDbsAnimeAltNames
	return dbsAnimeAltNamesInstance
}

func (d *PgDbsAnimeAltNames) AddAltName(reqAltName *AnimeAltName) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames AddAltName: Conn: %v", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: AnimeAltNames AddAltName: Rollback: %v", err)
		}
	}()

	if err := InsertAltNameWithNew(tx, reqAltName); err != nil {
		log.Printf("error: Dbs: AnimeAltNames AddAltName: InsertAltName: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("error: Dbs: AnimeAltNames AddAltName: Rollback: %v", err)
		return err
	}

	return nil
}

func (d *PgDbsAnimeAltNames) DeleteAltNames(reqAltNames []*AnimeAltName) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames DeleteAltNames: Conn: %v", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: AnimeAltNames DeleteAltNames: Rollback: %v", err)
		}
	}()

	if err := DeleteAltNames(tx, reqAltNames); err != nil {
		log.Printf("error: Dbs: AnimeAltNames DeleteAltNames: DeleteAltNames: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("error: Dbs: AnimeAltNames DeleteAltNames: Rollback: %v", err)
		return err
	}

	return nil
}

func InsertAltName(tx *sql.Tx, params *AnimeAltName) error {
	query := `INSERT INTO rel_anime_anime_names (id, anime_id, anime_names_id)
		VALUES ($1, $2, $3)`

	queryResult, err := tx.Exec(
		query,
		uuid.New(),
		params.AnimeId,
		params.AnimeName.Id,
	)
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltName: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltName: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltName: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}

func InsertAltNameWithNew(tx *sql.Tx, params *AnimeAltName) error {
	query := `
		WITH source_data(new_name_id, new_anime_name) AS (
			VALUES ($1::uuid, $2)
		), upsert AS (
			MERGE INTO anime_names
			USING source_data
				ON anime_names.name = source_data.new_anime_name
			WHEN MATCHED THEN
				UPDATE SET name = anime_names.name
			WHEN NOT MATCHED THEN
				INSERT (id, name)
				VALUES (source_data.new_name_id, source_data.new_anime_name)
			RETURNING id, created_at, updated_at, name
		)
		INSERT INTO rel_anime_anime_names (id, anime_id, anime_names_id)
		SELECT $3, $4, upsert.id
		FROM upsert
	`

	queryResult, err := tx.Exec(
		query,
		uuid.New(),
		params.AnimeName.Name,
		uuid.New(),
		params.AnimeId,
	)
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltNameWithNew: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltNameWithNew: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: AnimeAltNames: InsertAltNameWithNew: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}

func DeleteAltNames(tx *sql.Tx, reqAltNames []*AnimeAltName) error {
	query := `
	DELETE FROM rel_anime_anime_names alt_names
	WHERE alt_names.anime_id = $1
	AND alt_names.anime_names_id = ANY($2)
	`

	if len(reqAltNames) == 0 {
		return sql.ErrNoRows
	}

	animeId := reqAltNames[0].AnimeId
	args := make([]uuid.UUID, 0)
	for _, v := range reqAltNames {
		args = append(args, v.AnimeName.Id)
	}

	queryResult, err := tx.Exec(
		query,
		animeId,
		args,
	)
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: DeleteAltNames: Query: %v", err)
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: AnimeAltNames: DeleteAltNames: RowsAffected: %v", err)
	}

	if n == 0 {
		log.Printf("error: Dbs: AnimeAltNames: DeleteAltNames: RowsAffected: %v", n)
	}

	return nil
}

func SelectAllAnimeAltNames(tx *sql.Tx) ([]*AnimeAltName, error) {
	result := make([]*AnimeAltName, 0)

	query := `SELECT
	ran.id, ran.created_at, ran.updated_at, ran.anime_id,
	an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_anime_names ran
	JOIN anime_names an ON ran.anime_names_id = an.id`

	rows, err := tx.Query(query)
	if err != nil {
		log.Printf("error: DbsRelAnimeAnimeNames SelectAllNamesAnime: Query: %v", err)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: DbsRelAnimeAnimeNames SelectAllNamesAnime: Close rows: %v", err)
		}
	}()

	for rows.Next() == true {
		rel := &AnimeAltName{
			AnimeName: AnimeName{},
		}
		err := rows.Scan(
			&rel.Id,
			&rel.CreatedAt,
			&rel.UpdatedAt,
			&rel.AnimeId,
			&rel.AnimeName.Id,
			&rel.AnimeName.CreatedAt,
			&rel.AnimeName.UpdatedAt,
			&rel.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: DbsRelAnimeAnimeNames SelectAllNamesAnime: Scan: %v", err)
			return nil, err
		}
		result = append(result, rel)
	}

	return result, nil
}

func SelectAnimeAltNames(tx *sql.Tx, reqAnime []*Anime) ([]*AnimeAltName, error) {
	result := make([]*AnimeAltName, 0)

	args := make([]uuid.UUID, 0)
	for _, v := range reqAnime {
		args = append(args, v.Id)
	}

	if len(args) == 0 {
		return nil, sql.ErrNoRows
	}

	query := `SELECT
	ran.id, ran.created_at, ran.updated_at, ran.anime_id,
	an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_anime_names ran
	JOIN anime_names an ON ran.anime_names_id = an.id
	WHERE ran.anime_id = ANY($1)`

	queryRows, err := tx.Query(
		query,
		args,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeAnimeNames SelectAnimeNames: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() == true {
		rel := &AnimeAltName{
			AnimeName: AnimeName{},
		}
		err := queryRows.Scan(
			&rel.Id,
			&rel.CreatedAt,
			&rel.UpdatedAt,
			&rel.AnimeId,
			&rel.AnimeName.Id,
			&rel.AnimeName.CreatedAt,
			&rel.AnimeName.UpdatedAt,
			&rel.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: DbsRelAnimeAnimeNames SelectAnimeNames: Scan: %v", err)
			return nil, err
		}
		result = append(result, rel)
	}

	return result, nil
}
