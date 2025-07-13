package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type Anime struct {
	Id               uuid.UUID    `json:"id"`
	CreatedAt        time.Time    `json:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at"`
	Kind             string       `json:"kind"`
	Episodes         *int         `json:"episodes"`
	Description      *string      `json:"description"` // TODO: maybe remove this.
	ImageUrl         *string      `json:"image_url"`
	AnimeName        AnimeName    `json:"anime_name"`
	AlternativeNames []*AnimeName `json:"alternative_names"`
}

type DbsAnime interface {
	InsertAnime(anime *Anime) (*Anime, error)
	GetAnimeById(params *Anime) (*Anime, error)
	GetAllAnime() ([]*Anime, error)
	UpdateAnime(anime *Anime) error
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
	queryInsertAnime := `INSERT INTO anime (id, episodes, description, image_url, anime_names_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at, kind, episodes, description, image_url`

	err = tx.QueryRow(
		queryInsertAnime,
		uuid.New(),
		anime.Episodes,
		anime.Description,
		anime.ImageUrl,
		newAnime.AnimeName.Id,
	).Scan(
		&newAnime.Id,
		&newAnime.CreatedAt,
		&newAnime.UpdatedAt,
		&newAnime.Kind,
		&newAnime.Episodes,
		&newAnime.Description,
		&newAnime.ImageUrl,
	)
	if err != nil {
		log.Printf("insert anime")
		return nil, err
	}

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

func (d *PgDbsAnime) GetAnimeById(params *Anime) (*Anime, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsAnime GetAnimeById: Rollback: %v", err)
		}
	}()

	existingAnime, err := SelectAnimeJoinName(tx, params)
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: SelectAnimeJoinName: %v", err)
		return nil, err
	}

	altNames, err := SelectAllNamesByAnimeId(tx, params)
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: SelectAllNamesByAnimeId: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: Commit: %v", err)
		return nil, err
	}

	existingAnime.AlternativeNames = altNames

	return existingAnime, nil
}

func (d *PgDbsAnime) GetAllAnime() ([]*Anime, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsAnime GetAllAnime: Rollback: %v", err)
		}
	}()

	allAnime, err := SelectAllAnimeJoinName(tx)
	if err != nil {
		log.Printf("error: DbsAnime GetAllAnime: SelectAllAnimeJoinName: %v", err)
		return nil, err
	}

	allNames, err := SelectAllNamesAnime(tx)
	if err != nil {
		log.Printf("error: DbsAnime GetAllAnime: SelectAllNamesAnime: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnime GetAllAnime: Commit: %v", err)
		return nil, err
	}

	namesMap := make(map[uuid.UUID][]*AnimeName)
	for _, v := range allNames {
		curId := v.AnimeId

		_, ok := namesMap[curId]
		if !ok {
			namesMap[curId] = make([]*AnimeName, 0)
		}

		name := &AnimeName{
			Id:        v.AnimeName.Id,
			CreatedAt: v.AnimeName.CreatedAt,
			UpdatedAt: v.AnimeName.UpdatedAt,
			Name:      v.AnimeName.Name,
		}

		namesMap[curId] = append(namesMap[curId], name)
	}

	for k, v := range allAnime {
		curId := v.Id
		names, ok := namesMap[curId]
		if ok {
			allAnime[k].AlternativeNames = names
		}
	}

	return allAnime, nil
}

func (d *PgDbsAnime) UpdateAnime(anime *Anime) error {
	query := `UPDATE anime
	SET
		updated_at = $2,
		episodes = $3,
		description = $4,
		image_url = $5,
		anime_names_id = $6
	WHERE id = $1
	`

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: dbs anime UpdateAnime: %v", err)
		}
	}()

	result, err := tx.Exec(
		query,
		anime.Id,
		anime.UpdatedAt,
		anime.Episodes,
		anime.Description,
		anime.ImageUrl,
		anime.AnimeName.Id,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err == nil {
		if n == 0 {
			return sql.ErrNoRows
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func SelectAnimeJoinName(tx *sql.Tx, params *Anime) (*Anime, error) {
	existingAnime := &Anime{
		AnimeName: AnimeName{},
	}

	query := `SELECT a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url, an.id, an.created_at, an.updated_at, an.name
	FROM anime a JOIN anime_names an ON a.anime_names_id = an.id
	WHERE a.id = $1`

	err := tx.QueryRow(
		query,
		params.Id,
	).Scan(
		&existingAnime.Id,
		&existingAnime.CreatedAt,
		&existingAnime.UpdatedAt,
		&existingAnime.Kind,
		&existingAnime.Episodes,
		&existingAnime.Description,
		&existingAnime.ImageUrl,
		&existingAnime.AnimeName.Id,
		&existingAnime.AnimeName.CreatedAt,
		&existingAnime.AnimeName.UpdatedAt,
		&existingAnime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: DbsAnime SelectAnimeJoinName: Scan: %v", err)
		return nil, err
	}

	return existingAnime, nil
}

func SelectAllAnimeJoinName(tx *sql.Tx) ([]*Anime, error) {
	animeList := make([]*Anime, 0)

	query := `SELECT a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM anime a
	JOIN anime_names an ON a.anime_names_id = an.id`

	rows, err := tx.Query(query)
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("error: DbsAnime SelectAllAnimeJoinName: close rows: %v", err)
		}
	}()
	if err != nil {
		log.Printf("error: DbsAnime SelectAllAnimeJoinName: Query: %v", err)
		return nil, err
	}

	for rows.Next() == true {
		anime := &Anime{
			AnimeName: AnimeName{},
		}
		err := rows.Scan(
			&anime.Id,
			&anime.CreatedAt,
			&anime.UpdatedAt,
			&anime.Kind,
			&anime.Episodes,
			&anime.Description,
			&anime.ImageUrl,
			&anime.AnimeName.Id,
			&anime.AnimeName.CreatedAt,
			&anime.AnimeName.UpdatedAt,
			&anime.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: DbsAnime SelectAllAnimeJoinName: Scan: %v", err)
			return nil, err
		}
		animeList = append(animeList, anime)
	}

	return animeList, nil
}
