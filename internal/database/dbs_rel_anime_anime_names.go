package database

import (
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
	rels := make([]*RelAnimeAnimeNames, 0)

	query := `SELECT ran.id, ran.created_at, ran.updated_at, ran.anime_id, an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_anime_names ran
	JOIN anime_names an ON ran.anime_names_id = an.id`

	rows, err := d.db.Conn().Query(
		query,
	)
	if err != nil {
		log.Printf("error: dbsRelAnimeAnimeNames GetNames: query: %v", err)
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: dbsRelAnimeAnimeNames GetNames: close rows: %v", err)
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
			log.Printf("error: dbsRelAnimeAnimeNames GetNames: scan: %v", err)
			return nil, err
		}
		rels = append(rels, rel)
	}

	return rels, nil
}
