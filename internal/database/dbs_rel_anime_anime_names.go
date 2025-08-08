package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type RelAnimeAnimeNames struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	AnimeId   uuid.UUID `json:"anime_id"`
	AnimeName AnimeName `json:"anime_name"`
}

type DbsRelAnimeAnimeNames interface {
}

type PgDbsRelAnimeAnimeNames struct {
	db DbService
}

var dbsRelAnimeAnimeNamesInstance *PgDbsRelAnimeAnimeNames

func NewDbsRelAnimeAnimeNames(db DbService) DbsRelAnimeAnimeNames {
	if dbsRelAnimeAnimeNamesInstance != nil {
		return dbsRelAnimeAnimeNamesInstance
	}

	newDbsAnimeAnimeNames := &PgDbsRelAnimeAnimeNames{
		db: db,
	}
	dbsRelAnimeAnimeNamesInstance = newDbsAnimeAnimeNames
	return dbsRelAnimeAnimeNamesInstance
}

func InsertRelAnimeAnimeNames(tx *sql.Tx, params *RelAnimeAnimeNames) error {
	query := `INSERT INTO rel_anime_anime_names (id, anime_id, anime_names_id)
		VALUES ($1, $2, $3)`

	queryResult, err := tx.Exec(
		query,
		uuid.New(),
		params.AnimeId,
		params.AnimeName.Id,
	)
	if err != nil {
		log.Printf("error: Dbs: RelAnimeAnimeNames: InsertRelAnimeAnimeNames: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: RelAnimeAnimeNames: InsertRelAnimeAnimeNames: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: RelAnimeAnimeNames: InsertRelAnimeAnimeNames: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}

func SelectAllNamesAnime(tx *sql.Tx) ([]*RelAnimeAnimeNames, error) {
	result := make([]*RelAnimeAnimeNames, 0)

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
		rel := &RelAnimeAnimeNames{
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

func SelectAnimeNames(tx *sql.Tx, reqAnime []*Anime) ([]*RelAnimeAnimeNames, error) {
	result := make([]*RelAnimeAnimeNames, 0)

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
		rel := &RelAnimeAnimeNames{
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
