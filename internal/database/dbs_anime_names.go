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
	GetNamesByAnime(params *Anime) ([]*AnimeName, error)
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

func (d *PgDbsAnimeNames) GetNamesByAnime(params *Anime) ([]*AnimeName, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsAnimeNames GetNamesByAnime: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsAnimeNames GetNamesByAnime: Rollback: %v", err)
		}
	}()

	names, err := SelectAllNamesByAnimeId(tx, params)
	if err != nil {
		log.Printf("error: DbsAnimeNames GetNamesByAnime: SelectAllNamesByAnimeId: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnimeNames GetNamesByAnime: Commit: %v", err)
		return nil, err
	}

	return names, nil
}

func SelectAllNamesByAnimeId(tx *sql.Tx, params *Anime) ([]*AnimeName, error) {
	animeNames := make([]*AnimeName, 0)

	query := `SELECT an.id, an.created_at, an.updated_at, an.name FROM anime_names an
	JOIN rel_anime_anime_names ran ON an.id = ran.anime_names_id
	WHERE ran.anime_id = $1`

	rows, err := tx.Query(
		query,
		params.Id,
	)
	if err != nil {
		log.Printf("error: DbsAnimeNames SelectAllNamesByAnimeId: Query: %v", err)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: DbsAnimeNames SelectAllNamesByAnimeId: close rows: %v", err)
		}
	}()

	for rows.Next() == true {
		animeName := &AnimeName{}
		err := rows.Scan(
			&animeName.Id,
			&animeName.CreatedAt,
			&animeName.UpdatedAt,
			&animeName.Name,
		)
		if err != nil {
			log.Printf("error: DbsAnimeNames SelectAllNamesByAnimeId: Scan: %v", err)
			return nil, err
		}
		animeNames = append(animeNames, animeName)
	}

	return animeNames, nil
}
