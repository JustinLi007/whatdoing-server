package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/algo"
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
	InsertAnime(reqAnime *Anime) (*Anime, error)
	GetAnimeById(reqAnime *Anime) (*Anime, error)
	GetAllAnime(reqUser *User, opts ...OptionsFunc) ([]*Anime, error)
	UpdateAnime(reqAnime *Anime) error
	DeleteAnime(reqAnime *Anime) error
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

func (d *PgDbsAnime) InsertAnime(reqAnime *Anime) (*Anime, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: Anime: InsertAnime: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: Anime: InsertAnime: Rollback: %v", err)
		}
	}()

	dbAnimeName, err := InsertAnimeNameIfNotExist(tx, &reqAnime.AnimeName)
	if dbAnimeName != nil {
		log.Printf("error: Dbs: Anime: InsertAnime: SelectAltNameByAnimeName: %v", errors.New(fmt.Sprintf("duplicate record found: '%v'", dbAnimeName.Name)))
		return nil, err
	} else if err != nil && err != sql.ErrNoRows {
		log.Printf("error: Dbs: Anime: InsertAnime: SelectAltNameByAnimeName: %v", err)
		return nil, err
	}

	dbAnime, err := InsertAnime(tx, reqAnime)
	if err != nil {
		log.Printf("error: Dbs: Anime: InsertAnime: InsertAnime: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Anime: InsertAnime: Commit: %v", err)
		return nil, err
	}

	return dbAnime, nil
}

func (d *PgDbsAnime) GetAnimeById(reqAnime *Anime) (*Anime, error) {
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

	dbAnime, err := SelectAnimeJoinName(tx, reqAnime)
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: SelectAnimeJoinName: %v", err)
		return nil, err
	}

	temp := []*Anime{dbAnime}
	allNames, err := SelectAnimeNames(tx, temp)
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: SelectAllNamesByAnimeId: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsAnime GetAnimeById: Commit: %v", err)
		return nil, err
	}

	namesMap := buildNamesMap(allNames)
	if altNames, ok := namesMap[dbAnime.Id]; ok {
		dbAnime.AlternativeNames = altNames
	} else {
		dbAnime.AlternativeNames = make([]*AnimeName, 0)
	}

	return dbAnime, nil
}

func (d *PgDbsAnime) GetAllAnime(reqUser *User, opts ...OptionsFunc) ([]*Anime, error) {
	var err error

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

	options := NewOptions()
	for _, v := range opts {
		v(options)
	}

	orderBy := SORT_ASC
	if options.Sort != nil {
		orderBy = options.Sort.SortValue
	}

	var msg string
	animeList := make([]*Anime, 0)
	if options.IgnoreInLibrary && reqUser != nil {
		if animeList, err = SelectAnimeNotInLibrary(tx, reqUser, orderBy); err != nil {
			msg = fmt.Sprintf("error: Dbs: Anime: GetAllAnime: SelectAnimeNotInLibrary: %v", err)
		}
	} else {
		if animeList, err = SelectAllAnimeJoinName(tx, orderBy); err != nil {
			msg = fmt.Sprintf("error: Dbs: Anime: GetAllAnime: SelectAllAnimeJoinName: %v", err)
		}
	}
	if err != nil {
		log.Printf("%v", msg)
		return nil, err
	}

	allNames, err := SelectAnimeNames(tx, animeList)
	if err != nil {
		log.Printf("error: Dbs: Anime: GetAllAnime: SelectAnimeNames: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Anime: GetAllAnime: Commit: %v", err)
		return nil, err
	}

	namesMap := buildNamesMap(allNames)
	for k, v := range animeList {
		if names, ok := namesMap[v.Id]; ok {
			animeList[k].AlternativeNames = names
		}
	}

	if options.Search != nil {
		foundKmp := func(anime *Anime, targetString string) bool {
			idx := algo.Kmp(strings.ToLower(anime.AnimeName.Name), targetString)
			if idx == -1 {
				for _, v := range anime.AlternativeNames {
					if idx = algo.Kmp(strings.ToLower(v.Name), targetString); idx != -1 {
						break
					}
				}
			}
			return idx != -1
		}

		foundEditDistance := func(anime *Anime, targetString string, edits int) bool {
			result := algo.EditDistance(strings.ToLower(anime.AnimeName.Name), targetString)
			if result > edits {
				for _, v := range anime.AlternativeNames {
					result = algo.EditDistance(strings.ToLower(v.Name), targetString)
					if result <= edits {
						break
					}
				}
			}
			return result <= edits
		}

		filteredAnimeList := make([]*Anime, 0)
		kmpMatch := false
		editDistanceMatch := false
		for _, v := range animeList {
			kmpMatch = foundKmp(v, options.Search.SearchValue)
			editDistanceMatch = foundEditDistance(v, options.Search.SearchValue, 2)
			if kmpMatch || editDistanceMatch {
				filteredAnimeList = append(filteredAnimeList, v)
			}
		}
		animeList = filteredAnimeList
	}

	return animeList, nil
}

func (d *PgDbsAnime) UpdateAnime(reqAnime *Anime) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsAnime UpdateAnime: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsAnime UpdateAnime: Rollback: %v", err)
		}
	}()

	err = UpdateAnimeById(tx, reqAnime)
	if err != nil {
		log.Printf("error: DbsAnime UpdateAnime: UpdateAnimeById: %v", err)
		return err
	}

	for _, v := range reqAnime.AlternativeNames {
		_, err := SelectAnimeNameByName(tx, v)
		if err != nil {
			dbAnimeName, err := InsertAnimeName(tx, v)
			if err != nil {
				log.Printf("error: DbsAnime UpdateAnime: InsertAnimeName: %v", err)
				return err
			}

			relReq := &RelAnimeAnimeNames{
				AnimeId:   reqAnime.Id,
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

func (d *PgDbsAnime) DeleteAnime(reqAnime *Anime) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: Anime: DeleteAnime: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: Anime: DeleteAnime: Rollback: %v", err)
		}
	}()

	err = DeleteAnime(tx, reqAnime)
	if err != nil {
		log.Printf("error: Dbs: Anime: DeleteAnime: DeleteAnime: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Anime: DeleteAnime: Commit: %v", err)
		return err
	}

	return nil
}

func InsertAnime(tx *sql.Tx, params *Anime) (*Anime, error) {
	result := &Anime{
		AnimeName:        AnimeName{},
		AlternativeNames: make([]*AnimeName, 0),
	}

	query := `
	WITH select_name AS (
		SELECT an.id, an.created_at, an.updated_at, an.name
		FROM anime_names an
		WHERE an.name = $5
	), insert_anime AS (
		INSERT INTO anime (id, episodes, description, image_url, anime_names_id)
		SELECT $1, $2, $3, $4, select_name.id
		FROM select_name
		RETURNING id, created_at, updated_at, kind, episodes, description, image_url, anime_names_id
	), insert_alt_name AS (
		INSERT INTO rel_anime_anime_names (id, anime_id, anime_names_id)
		SELECT $6, insert_anime.id, insert_anime.anime_names_id
		FROM insert_anime
	)
	SELECT
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM insert_anime a
	JOIN select_name an ON a.anime_names_id = an.id
	`

	err := tx.QueryRow(
		query,
		uuid.New(),
		params.Episodes,
		params.Description,
		params.ImageUrl,
		params.AnimeName.Name,
		uuid.New(),
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Kind,
		&result.Episodes,
		&result.Description,
		&result.ImageUrl,
		&result.AnimeName.Id,
		&result.AnimeName.CreatedAt,
		&result.AnimeName.UpdatedAt,
		&result.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: Dbs: Anime: InsertAnime: Query: %v", err)
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

func SelectAllAnimeJoinName(tx *sql.Tx, orderBy string) ([]*Anime, error) {
	animeList := make([]*Anime, 0)

	query := fmt.Sprintf(`
	SELECT a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM anime a
	JOIN anime_names an ON a.anime_names_id = an.id
	ORDER BY an.name %s
	`,
		orderBy,
	)

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

func SelectAnimeNotInLibrary(tx *sql.Tx, reqUser *User, orderBy string) ([]*Anime, error) {
	result := make([]*Anime, 0)

	query := fmt.Sprintf(`
	WITH user_lib AS (
		SELECT user_library.id FROM user_library WHERE user_id = $1
	)
	SELECT a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM anime a
	JOIN anime_names an ON a.anime_names_id = an.id
	WHERE a.id NOT IN (
		SELECT progress.anime_id
		FROM progress_anime progress,
		user_lib
		WHERE progress.user_library_id = user_lib.id
	)
	ORDER BY an.name %s
	`,
		orderBy,
	)

	queryRows, err := tx.Query(
		query,
		reqUser.Id,
	)
	if err != nil {
		log.Printf("error: Dbs: Anime: SelectAnimeInLibrary: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() == true {
		temp := &Anime{
			AnimeName:        AnimeName{},
			AlternativeNames: make([]*AnimeName, 0),
		}

		err := queryRows.Scan(
			&temp.Id,
			&temp.CreatedAt,
			&temp.UpdatedAt,
			&temp.Kind,
			&temp.Episodes,
			&temp.Description,
			&temp.ImageUrl,
			&temp.AnimeName.Id,
			&temp.AnimeName.CreatedAt,
			&temp.AnimeName.UpdatedAt,
			&temp.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: Dbs: Anime: SelectAnimeInLibrary: Scan: %v", err)
			return nil, err
		}

		result = append(result, temp)
	}

	return result, nil
}

func DeleteAnime(tx *sql.Tx, reqAnime *Anime) error {
	query := `
	DELETE FROM anime
	WHERE anime.id = $1
	`

	queryResult, err := tx.Exec(query)
	if err != nil {
		log.Printf("error: Dbs: Anime: DeleteAnime: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: Anime: DeleteAnime: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: Anime: DeleteAnime: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}
