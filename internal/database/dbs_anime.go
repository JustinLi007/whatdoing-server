package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type Anime struct {
	Id          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Episodes    *int      `json:"episodes"`
	Description *string   `json:"description"` // TODO: maybe remove this.
	AnimeName   AnimeName `json:"anime_name"`
}

type AnimeName struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Name      string    `json:"name"`
}

type DbsAnime interface {
	InsertAnime(anime *Anime) (*Anime, error)
	GetAnimeById(anime *Anime) (*Anime, error)
	GetAllAnime() ([]*Anime, error)
}

type PgDbsAnime struct {
	db DbService
}

var dbsAnimeInstance *PgDbsAnime

func NewDbsAnime(db DbService) DbsAnime {
	if dbsAnimeInstance != nil {
		return dbsAnimeInstance
	}

	newDbsAnime := &PgDbsAnime{
		db: db,
	}
	dbsAnimeInstance = newDbsAnime

	return dbsAnimeInstance
}

func (d *PgDbsAnime) InsertAnime(anime *Anime) (*Anime, error) {
	// TODO: this probably should return an err indicating already exist so caller can handle accordingly...

	newAnime := &Anime{
		AnimeName: AnimeName{},
	}
	newNameRel := false

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: dbs anime InsertAnime: %v", err)
		}
	}()

	queryGetAnimeName := `SELECT * FROM anime_names
	WHERE name = $1`

	err = tx.QueryRow(
		queryGetAnimeName,
		anime.AnimeName.Name,
	).Scan(
		&newAnime.AnimeName.Id,
		&newAnime.AnimeName.CreatedAt,
		&newAnime.AnimeName.UpdatedAt,
		&newAnime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("couldnt get anime name")
		queryInsertAnimeName := `INSERT INTO anime_names (id, name)
	VALUES ($1, $2)
	RETURNING id, created_at, updated_at, name`

		err = tx.QueryRow(
			queryInsertAnimeName,
			uuid.New(),
			anime.AnimeName.Name,
		).Scan(
			&newAnime.AnimeName.Id,
			&newAnime.AnimeName.CreatedAt,
			&newAnime.AnimeName.UpdatedAt,
			&newAnime.AnimeName.Name,
		)
		if err != nil {
			log.Printf("insert anime name")
			return nil, err
		}
		newNameRel = true
	}

	// TODO: include other fields?
	queryInsertAnime := `INSERT INTO anime (id, episodes, description, anime_names_id)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, updated_at, episodes, description`

	err = tx.QueryRow(
		queryInsertAnime,
		uuid.New(),
		anime.Episodes,
		anime.Description,
		newAnime.AnimeName.Id,
	).Scan(
		&newAnime.Id,
		&newAnime.CreatedAt,
		&newAnime.UpdatedAt,
		&newAnime.Episodes,
		&newAnime.Description,
	)
	if err != nil {
		log.Printf("insert anime")
		return nil, err
	}

	log.Printf("before here...")

	if newNameRel {
		queryInsertRelAnimeAnimeName := `INSERT INTO rel_anime_anime_names (id, anime_id, anime_names_id)
		VALUES ($1, $2, $3)`

		result, err := tx.Exec(
			queryInsertRelAnimeAnimeName,
			uuid.New(),
			newAnime.Id,
			newAnime.AnimeName.Id,
		)
		if err != nil {
			log.Printf("insert rel")
			return nil, err
		}
		n, err := result.RowsAffected()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, sql.ErrNoRows
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("commit")
		return nil, err
	}

	return newAnime, nil
}

func (d *PgDbsAnime) GetAnimeById(anime *Anime) (*Anime, error) {
	existingAnime := &Anime{
		AnimeName: AnimeName{},
	}

	query := `SELECT a.*, an.* FROM anime a
	JOIN anime_names an ON a.anime_names_id = an.id
	WHERE a.id = $1`

	err := d.db.Conn().QueryRow(
		query,
		anime.Id,
	).Scan(
		&existingAnime.Id,
		&existingAnime.CreatedAt,
		&existingAnime.UpdatedAt,
		&existingAnime.Episodes,
		&existingAnime.Description,
		&existingAnime.AnimeName.Id,
		&existingAnime.AnimeName.CreatedAt,
		&existingAnime.AnimeName.UpdatedAt,
		&existingAnime.AnimeName.Name,
	)
	if err != nil {
		return nil, err
	}

	return existingAnime, nil
}

func (d *PgDbsAnime) GetAllAnime() ([]*Anime, error) {
	animeList := make([]*Anime, 0)

	query := `SELECT a.id, a.created_at, a.updated_at, a.episodes, a.description, an.id, an.created_at, an.updated_at, an.name FROM anime a
	JOIN anime_names an ON a.anime_names_id = an.id`

	rows, err := d.db.Conn().Query(
		query,
	)
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: dbs anime GetAllAnime: close rows: %v", err)
		}
	}()
	if err != nil {
		return nil, err
	}

	for rows.Next() == true {
		anime := &Anime{}
		err := rows.Scan(
			&anime.Id,
			&anime.CreatedAt,
			&anime.UpdatedAt,
			&anime.Episodes,
			&anime.Description,
			&anime.AnimeName.Id,
			&anime.AnimeName.CreatedAt,
			&anime.AnimeName.UpdatedAt,
			&anime.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: dbs anime GetAllAnime scan: %v", err)
			return nil, err
		}
		animeList = append(animeList, anime)
	}

	return animeList, nil
}
