package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

type RelAnimeUserLibrary struct {
	Id            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Status        string    `json:"status"`
	Episode       int       `json:"episode"`
	Anime         *Anime    `json:"anime"`
	UserLibraryId uuid.UUID `json:"-"`
}

const (
	STARTED     = "started"
	NOT_STARTED = "not-started"
	COMPLETED   = "completed"
)

type options struct {
	relId       uuid.UUID
	animeId     uuid.UUID
	status      string
	withRelId   bool
	withAnimeId bool
}

func newOptions() *options {
	options := &options{
		status:    "",
		withRelId: false,
	}
	return options
}

type OptionsFunc func(o *options)

func WithRelId(id uuid.UUID) OptionsFunc {
	return func(o *options) {
		o.withRelId = true
		o.relId = id
	}
}

func WithAnimeId(id uuid.UUID) OptionsFunc {
	return func(o *options) {
		o.withAnimeId = true
		o.animeId = id
	}
}

func WithStatus(status string) OptionsFunc {
	return func(o *options) {
		switch status {
		case STARTED:
			o.status = STARTED
		case NOT_STARTED:
			o.status = NOT_STARTED
		case COMPLETED:
			o.status = COMPLETED
		default:
			o.status = ""
		}
	}
}

type DbsRelAnimeUserLibrary interface {
	AddToLibrary(reqUser *User, reqAnime *Anime) (*RelAnimeUserLibrary, error)
	UpdateProgress(reqUser *User, reqRelAnimeUserLibrary *RelAnimeUserLibrary) (*RelAnimeUserLibrary, error)
	GetProgress(reqUser *User, opts ...OptionsFunc) ([]*RelAnimeUserLibrary, error)
}

type PgDbsRelAnimeUserLibrary struct {
	db DbService
}

var dbsRelAnimeUserLibraryInstance *PgDbsRelAnimeUserLibrary

func NewDbsRelAnimeUserLibrary(db DbService) DbsRelAnimeUserLibrary {
	if dbsRelAnimeUserLibraryInstance != nil {
		return dbsRelAnimeUserLibraryInstance
	}
	newDbsRelAnimeUserLibrary := &PgDbsRelAnimeUserLibrary{
		db: db,
	}
	dbsRelAnimeUserLibraryInstance = newDbsRelAnimeUserLibrary

	return dbsRelAnimeUserLibraryInstance
}

func (d *PgDbsRelAnimeUserLibrary) AddToLibrary(reqUser *User, reqAnime *Anime) (*RelAnimeUserLibrary, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: Rollback: %v", err)
		}
	}()

	dbRelAnimeUserLibrary, err := InsertRelAnimeUserLibrary(tx, reqUser, reqAnime)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: InsertRelAnimeUserLibrary: %v", err)
		return nil, err
	}

	temp := []*Anime{dbRelAnimeUserLibrary.Anime}
	allNames, err := SelectAnimeNames(tx, temp)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: SelectAllNamesByAnimeId: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: Commit: %v", err)
		return nil, err
	}

	namesMap := buildNamesMap(allNames)
	if altNames, ok := namesMap[dbRelAnimeUserLibrary.Anime.Id]; ok {
		dbRelAnimeUserLibrary.Anime.AlternativeNames = altNames
	} else {
		dbRelAnimeUserLibrary.Anime.AlternativeNames = make([]*AnimeName, 0)
	}

	return dbRelAnimeUserLibrary, nil
}

func (d *PgDbsRelAnimeUserLibrary) UpdateProgress(reqUser *User, reqRelAnimeUserLibrary *RelAnimeUserLibrary) (*RelAnimeUserLibrary, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: Rollback: %v", err)
		}
	}()

	dbRelAnimeUserLibrary, err := UpdateRelAnimeUserLibrary(tx, reqUser, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: UpdateRelAnimeUserLibrary: %v", err)
		return nil, err
	}

	temp := []*Anime{dbRelAnimeUserLibrary.Anime}
	allNames, err := SelectAnimeNames(tx, temp)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: SelectAnimeNames: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: Commit: %v", err)
		return nil, err
	}

	namesMap := buildNamesMap(allNames)
	if altNames, ok := namesMap[dbRelAnimeUserLibrary.Anime.Id]; ok {
		dbRelAnimeUserLibrary.Anime.AlternativeNames = altNames
	}

	return dbRelAnimeUserLibrary, nil
}

func (d *PgDbsRelAnimeUserLibrary) GetProgress(reqUser *User, opts ...OptionsFunc) ([]*RelAnimeUserLibrary, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary GetProgress: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsRelAnimeUserLibrary GetProgress: Rollback: %v", err)
		}
	}()

	options := newOptions()
	for _, v := range opts {
		v(options)
	}

	result := make([]*RelAnimeUserLibrary, 0)
	if options.withRelId {
		reqRelAnimeUserLibrary := &RelAnimeUserLibrary{
			Id: options.relId,
		}
		dbRelAnimeUserLibrary, err := SelectRelAnimeUserLibraryById(tx, reqUser, reqRelAnimeUserLibrary)
		if err != nil {
			log.Printf("error: DbsRelAnimeUserLibrary GetProgress: SelectRelAnimeUserLibraryById: %v", err)
			return nil, err
		}
		result = append(result, dbRelAnimeUserLibrary)
	} else if options.withAnimeId {
		reqAnime := &Anime{
			Id: options.animeId,
		}
		dbRelAnimeUserLibrary, err := SelectRelAnimeUserLibraryByAnimeId(tx, reqUser, reqAnime)
		if err != nil {
			log.Printf("error: DbsRelAnimeUserLibrary GetProgress: SelectRelAnimeUserLibraryByAnimeId: %v", err)
			return nil, err
		}
		result = append(result, dbRelAnimeUserLibrary)
	} else if options.status != "" {
		reqRelAnimeUserLibrary := &RelAnimeUserLibrary{
			Status: options.status,
		}
		dbRelAnimeUserLibrary, err := SelectRelAnimeUserLibraryByStatus(tx, reqUser, reqRelAnimeUserLibrary)
		if err != nil {
			log.Printf("error: DbsRelAnimeUserLibrary GetProgress: SelectRelAnimeUserLibraryByStatus: %v", err)
			return nil, err
		}
		result = dbRelAnimeUserLibrary
	} else {
		msg := "error: DbsRelAnimeUserLibrary GetProgress: invalid options"
		log.Printf(msg)
		return nil, errors.New(msg)
	}

	allNames := make([]*RelAnimeAnimeNames, 0)
	if len(result) > 0 {
		tempAnime := make([]*Anime, 0)
		for _, v := range result {
			tempAnime = append(tempAnime, v.Anime)
		}

		allNames, err = SelectAnimeNames(tx, tempAnime)
		if err != nil {
			log.Printf("error: DbsRelAnimeUserLibrary GetProgress: SelectAnimeNames: %v", err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary GetProgress: Commit: %v", err)
		return nil, err
	}

	namesMap := buildNamesMap(allNames)

	for k, v := range result {
		curId := v.Anime.Id
		if altNames, ok := namesMap[curId]; ok {
			result[k].Anime.AlternativeNames = altNames
		}
	}

	return result, nil
}

func InsertRelAnimeUserLibrary(tx *sql.Tx, reqUser *User, reqAnime *Anime) (*RelAnimeUserLibrary, error) {
	result := &RelAnimeUserLibrary{
		Anime: &Anime{},
	}

	query := `WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	),
	new_rel AS (
		INSERT INTO rel_anime_user_library (id, anime_id, user_library_id)
		SELECT $2, $3, user_lib.id
		FROM user_lib
		RETURNING id, created_at, updated_at, status, episode, anime_id
	)
	SELECT new_rel.id, new_rel.created_at, new_rel.updated_at, new_rel.status, new_rel.episode,
	anime.id, anime.created_at, anime.updated_at, anime.kind, anime.episodes, anime.description, anime.image_url,
	anime_names.id, anime_names.created_at, anime_names.updated_at, anime_names.name
	FROM new_rel
	JOIN anime ON anime.id = new_rel.anime_id
	JOIN anime_names ON anime.anime_names_id = anime_names.id`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		uuid.New(),
		reqAnime.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Status,
		&result.Episode,
		&result.Anime.Id,
		&result.Anime.CreatedAt,
		&result.Anime.UpdatedAt,
		&result.Anime.Kind,
		&result.Anime.Episodes,
		&result.Anime.Description,
		&result.Anime.ImageUrl,
		&result.Anime.AnimeName.Id,
		&result.Anime.AnimeName.CreatedAt,
		&result.Anime.AnimeName.UpdatedAt,
		&result.Anime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary InsertRelAnimeUserLibrary: Query: %v", err)
		return nil, err
	}

	return result, nil
}

func UpdateRelAnimeUserLibrary(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *RelAnimeUserLibrary) (*RelAnimeUserLibrary, error) {
	result := &RelAnimeUserLibrary{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	),
	updated_rel AS (
		UPDATE rel_anime_user_library
		SET
			updated_at = $3,
			status = $4,
			episode = $5
		FROM user_lib
		WHERE id = $2 AND user_library_id = user_lib.id
		RETURNING id, created_at, updated_at, status, episode, anime_id
	)
	SELECT
	updated_rel.id, updated_rel.created_at, updated_rel.updated_at, updated_rel.status, updated_rel.episode,
	anime.id, anime.created_at, anime.updated_at, anime.kind, anime.episodes, anime.description, anime.image_url,
	anime_names.id, anime_names.created_at, anime_names.updated_at, anime_names.name
	FROM updated_rel
	JOIN anime ON updated_rel.anime_id = anime.id
	JOIN anime_names ON anime.anime_names_id = anime_names.id
	`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Id,
		time.Now(),
		reqRelAnimeUserLibrary.Status,
		reqRelAnimeUserLibrary.Episode,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Status,
		&result.Episode,
		&result.Anime.Id,
		&result.Anime.CreatedAt,
		&result.Anime.UpdatedAt,
		&result.Anime.Kind,
		&result.Anime.Episodes,
		&result.Anime.Description,
		&result.Anime.ImageUrl,
		&result.Anime.AnimeName.Id,
		&result.Anime.AnimeName.CreatedAt,
		&result.Anime.AnimeName.UpdatedAt,
		&result.Anime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateRelAnimeUserLibrary: Query: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectRelAnimeUserLibraryByAnimeId(tx *sql.Tx, reqUser *User, reqAnime *Anime) (*RelAnimeUserLibrary, error) {
	result := &RelAnimeUserLibrary{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	)
	SELECT
	ul.id, ul.created_at, ul.updated_at, ul.status, ul.episode,
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_user_library ul
	JOIN anime a ON ul.anime_id = a.id
	JOIN anime_names an ON a.anime_names_id = an.id
	JOIN user_lib ON ul.user_library_id = user_lib.id
	WHERE a.id = $2
	`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		reqAnime.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Status,
		&result.Episode,
		&result.Anime.Id,
		&result.Anime.CreatedAt,
		&result.Anime.UpdatedAt,
		&result.Anime.Kind,
		&result.Anime.Episodes,
		&result.Anime.Description,
		&result.Anime.ImageUrl,
		&result.Anime.AnimeName.Id,
		&result.Anime.AnimeName.CreatedAt,
		&result.Anime.AnimeName.UpdatedAt,
		&result.Anime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary SelectRelAnimeUserLibraryByAnimeId: Scan: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectRelAnimeUserLibraryById(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *RelAnimeUserLibrary) (*RelAnimeUserLibrary, error) {
	result := &RelAnimeUserLibrary{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	)
	SELECT
	ul.id, ul.created_at, ul.updated_at, ul.status, ul.episode,
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_user_library ul
	JOIN anime a ON ul.anime_id = a.id
	JOIN anime_names an ON a.anime_names_id = an.id
	JOIN user_lib ON ul.user_library_id = user_lib.id
	WHERE ul.id = $2
	`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Status,
		&result.Episode,
		&result.Anime.Id,
		&result.Anime.CreatedAt,
		&result.Anime.UpdatedAt,
		&result.Anime.Kind,
		&result.Anime.Episodes,
		&result.Anime.Description,
		&result.Anime.ImageUrl,
		&result.Anime.AnimeName.Id,
		&result.Anime.AnimeName.CreatedAt,
		&result.Anime.AnimeName.UpdatedAt,
		&result.Anime.AnimeName.Name,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary SelectRelAnimeUserLibraryById: Scan: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectRelAnimeUserLibraryByStatus(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *RelAnimeUserLibrary) ([]*RelAnimeUserLibrary, error) {
	result := make([]*RelAnimeUserLibrary, 0)

	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	)
	SELECT
	ul.id, ul.created_at, ul.updated_at, ul.status, ul.episode,
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM rel_anime_user_library ul
	JOIN anime a ON ul.anime_id = a.id
	JOIN anime_names an ON a.anime_names_id = an.id
	JOIN user_lib ON ul.user_library_id = user_lib.id
	WHERE ul.status = $2
	`

	queryRows, err := tx.Query(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Status,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary SelectRelAnimeUserLibraryByStatus: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() == true {
		rel := &RelAnimeUserLibrary{
			Anime: &Anime{
				AlternativeNames: make([]*AnimeName, 0),
			},
		}

		err := queryRows.Scan(
			&rel.Id,
			&rel.CreatedAt,
			&rel.UpdatedAt,
			&rel.Status,
			&rel.Episode,
			&rel.Anime.Id,
			&rel.Anime.CreatedAt,
			&rel.Anime.UpdatedAt,
			&rel.Anime.Kind,
			&rel.Anime.Episodes,
			&rel.Anime.Description,
			&rel.Anime.ImageUrl,
			&rel.Anime.AnimeName.Id,
			&rel.Anime.AnimeName.CreatedAt,
			&rel.Anime.AnimeName.UpdatedAt,
			&rel.Anime.AnimeName.Name,
		)
		if err != nil {
			log.Printf("error: DbsRelAnimeUserLibrary SelectRelAnimeUserLibraryByStatus: Scan: %v", err)
			return nil, err
		}

		result = append(result, rel)
	}

	return result, nil
}
