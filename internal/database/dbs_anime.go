package database

import (
	"database/sql"
	"errors"
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
	Description      *string      `json:"description"`
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
	result := &Anime{
		AnimeName: AnimeName{},
	}
	newNameRel := false

	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsAnime: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: DbsAnime InsertAnime: Rollback: %v", err)
		}
	}()

	dbAnimeName, err := SelectAnimeNameByName(tx, &anime.AnimeName)
	if err != nil {
		dbAnimeName, err = InsertAnimeName(tx, &anime.AnimeName)
		if err != nil {
			log.Printf("error: DbsAnime InsertAnime: InsertAnimeName: %v", err)
			return nil, err
		}
		newNameRel = true
	} else {
		// TODO: this probably should return an err indicating already exist so caller can handle accordingly...
		return nil, errors.New("error: already exists")
	}

	anime.AnimeName = *dbAnimeName
	dbAnime, err := InsertAnime(tx, anime)
	if err != nil {
		log.Printf("error: DbsAnime InsertAnime: InsertAnime: %v", err)
		return nil, err
	}

	result = dbAnime
	result.AnimeName = *dbAnimeName

	if newNameRel {
		relReq := &RelAnimeAnimeNames{
			AnimeId:   result.Id,
			AnimeName: result.AnimeName,
		}
		err := InsertRelAnimeAnimeNames(tx, relReq)
		if err != nil {
			log.Printf("error: DbsAnime InsertRelAnimeAnimeNames: %v:", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnime InsertAnime: Commit: %v", err)
		return nil, err
	}

	return result, nil
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
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsAnime UpdateAnime: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: DbsAnime UpdateAnime: Rollback: %v", err)
		}
	}()

	err = UpdateAnimeById(tx, anime)
	if err != nil {
		log.Printf("error: DbsAnime UpdateAnime: UpdateAnimeById: %v", err)
		return err
	}

	for _, v := range anime.AlternativeNames {
		_, err := SelectAnimeNameByName(tx, v)
		if err != nil {
			dbAnimeName, err := InsertAnimeName(tx, v)
			if err != nil {
				log.Printf("error: DbsAnime UpdateAnime: InsertAnimeName: %v", err)
				return err
			}

			relReq := &RelAnimeAnimeNames{
				AnimeId:   anime.Id,
				AnimeName: *dbAnimeName,
			}
			err = InsertRelAnimeAnimeNames(tx, relReq)
			if err != nil {
				log.Printf("error: DbsAnime UpdateAnime: InsertRelAnimeAnimeNames: %v", err)
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnime UpdateAnime: Commit: %v", err)
		return err
	}

	return nil
}

func InsertAnime(tx *sql.Tx, params *Anime) (*Anime, error) {
	result := &Anime{}

	query := `INSERT INTO anime (id, episodes, description, image_url, anime_names_id)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, created_at, updated_at, kind, episodes, description, image_url`

	err := tx.QueryRow(
		query,
		uuid.New(),
		params.Episodes,
		params.Description,
		params.ImageUrl,
		params.AnimeName.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Kind,
		&result.Episodes,
		&result.Description,
		&result.ImageUrl,
	)
	if err != nil {
		log.Printf("error: DbsAnime InsertAnime: Query: %v", err)
		return nil, err
	}

	return result, nil
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

func UpdateAnimeById(tx *sql.Tx, params *Anime) error {
	query := `UPDATE anime
	SET
		updated_at = $2,
		episodes = $3,
		description = $4,
		image_url = $5,
		anime_names_id = $6
	WHERE id = $1`

	queryResult, err := tx.Exec(
		query,
		params.Id,
		params.UpdatedAt,
		params.Episodes,
		params.Description,
		params.ImageUrl,
		params.AnimeName.Id,
	)
	if err != nil {
		return err
	}

	n, err := queryResult.RowsAffected()
	if err == nil {
		if n == 0 {
			log.Printf("error: DbsAnime UpdateAnimeById: RowsAffected: 0")
			return sql.ErrNoRows
		}
	}

	return nil
}
