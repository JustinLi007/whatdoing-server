package database

import (
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
	GetNamesByAnime(anime *Anime) ([]*AnimeName, error)
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

func (d *PgDbsAnimeNames) GetNamesByAnime(anime *Anime) ([]*AnimeName, error) {
	animeNames := make([]*AnimeName, 0)

	query := `SELECT an.id, an.created_at, an.updated_at, an.name FROM anime_names an
	JOIN rel_anime_anime_names ran ON an.id = ran.anime_names_id
	WHERE ran.anime_id = $1`

	rows, err := d.db.Conn().Query(
		query,
		anime.Id,
	)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: DbsAnimeNames GetNames: close rows: %v", err)
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
			log.Printf("error: DbsAnimeNames GetNames: scan: %v", err)
			return nil, err
		}
		animeNames = append(animeNames, animeName)
	}

	return animeNames, nil
}
