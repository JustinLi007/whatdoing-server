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
	GetNames() ([]*RelAnimeAnimeNames, error)
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

func (d *PgDbsRelAnimeAnimeNames) GetNames() ([]*RelAnimeAnimeNames, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsRelAnimeAnimeNames GetNames: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsRelAnimeAnimeNames GetNames: Rollback: %v", err)
		}
	}()

	names, err := SelectAllNamesAnime(tx)
	if err != nil {
		log.Printf("error: DbsRelAnimeAnimeNames SelectAllNamesAnime: Commit: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsRelAnimeAnimeNames GetNames: Commit: %v", err)
		return nil, err
	}

	return names, nil
}

func SelectAllNamesAnime(tx *sql.Tx) ([]*RelAnimeAnimeNames, error) {
	rels := make([]*RelAnimeAnimeNames, 0)

	query := `SELECT ran.id, ran.created_at, ran.updated_at, ran.anime_id,
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
		rels = append(rels, rel)
	}

	return rels, nil
}
