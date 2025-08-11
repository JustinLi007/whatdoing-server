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

type ProgressAnime struct {
	Id            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Episode       int       `json:"episode"`
	Anime         *Anime    `json:"anime"`
	UserLibraryId uuid.UUID `json:"-"`
}

type DbsProgressAnime interface {
	AddToLibrary(reqUser *User, reqAnime *Anime) (*ProgressAnime, error)
	UpdateProgress(reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) (*ProgressAnime, error)
	GetProgress(reqUser *User, opts ...OptionsFunc) ([]*ProgressAnime, error)
	RemoveProgress(reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) error
}

type PgDbsProgressAnime struct {
	db DbService
}

var dbsProgressAnimeInstance *PgDbsProgressAnime

func NewDbsProgressAnime(db DbService) DbsProgressAnime {
	if dbsProgressAnimeInstance != nil {
		return dbsProgressAnimeInstance
	}
	newDbsProgressAnime := &PgDbsProgressAnime{
		db: db,
	}
	dbsProgressAnimeInstance = newDbsProgressAnime

	return dbsProgressAnimeInstance
}

func (d *PgDbsProgressAnime) AddToLibrary(reqUser *User, reqAnime *Anime) (*ProgressAnime, error) {
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

	dbRelAnimeUserLibrary, err := InsertProgress(tx, reqUser, reqAnime)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: InsertProgress: %v", err)
		return nil, err
	}

	temp := []*Anime{dbRelAnimeUserLibrary.Anime}
	allNames, err := SelectAnimeNames(tx, temp)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary AddToLibrary: SelectAnimeNames: %v", err)
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

func (d *PgDbsProgressAnime) UpdateProgress(reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) (*ProgressAnime, error) {
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

	dbRelAnimeUserLibrary, err := UpdateProgress(tx, reqUser, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary UpdateProgress: UpdateProgress: %v", err)
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

func (d *PgDbsProgressAnime) GetProgress(reqUser *User, opts ...OptionsFunc) ([]*ProgressAnime, error) {
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

	options := NewOptions()
	for _, v := range opts {
		v(options)
	}

	orderBy := SORT_ASC
	if options.Sort != nil {
		orderBy = options.Sort.SortValue
	}

	result := make([]*ProgressAnime, 0)

	if options.ProgressId != nil || options.AnimeId != nil {
		var msg string
		var err error
		var dbProgress *ProgressAnime
		if options.ProgressId != nil {
			reqRelAnimeUserLibrary := &ProgressAnime{
				Id: options.ProgressId.Id,
			}
			dbProgress, err = SelectProgressById(tx, reqUser, reqRelAnimeUserLibrary)
			msg = fmt.Sprintf("error: DbsRelAnimeUserLibrary GetProgress: SelectRelAnimeUserLibraryById: %v", err)
		}
		if dbProgress == nil && options.AnimeId != nil {
			reqAnime := &Anime{
				Id: options.AnimeId.Id,
			}
			dbProgress, err = SelectProgressByAnimeId(tx, reqUser, reqAnime)
			msg = fmt.Sprintf("error: DbsRelAnimeUserLibrary GetProgress: SelectRelAnimeUserLibraryByAnimeId: %v", err)
		}
		if err != nil {
			log.Println(msg)
			return nil, err
		}
		result = append(result, dbProgress)
	} else if options.Status != nil {
		var msg string
		var err error
		var dbProgress []*ProgressAnime
		switch options.Status.StatusValue {
		case STATUS_NOT_STARTED:
			dbProgress, err = SelectProgressNotStarted(tx, reqUser, orderBy)
			msg = fmt.Sprintf("error: DbsRelAnimeUserLibrary GetProgress: SelectProgressNotStarted: %v", err)
		case STATUS_STARTED:
			dbProgress, err = SelectProgressStarted(tx, reqUser, orderBy)
			msg = fmt.Sprintf("error: DbsRelAnimeUserLibrary GetProgress: SelectProgressStarted: %v", err)
		case STATUS_COMPLETED:
			dbProgress, err = SelectProgressCompleted(tx, reqUser, orderBy)
			msg = fmt.Sprintf("error: DbsRelAnimeUserLibrary GetProgress: SelectProgressCompleted: %v", err)
		}
		if err != nil {
			log.Println(msg)
			return nil, err
		}
		result = dbProgress
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

		filteredProgressList := make([]*ProgressAnime, 0)
		kmpMatch := false
		editDistanceMatch := false

		for _, v := range result {
			kmpMatch = foundKmp(v.Anime, options.Search.SearchValue)
			editDistanceMatch = foundEditDistance(v.Anime, options.Search.SearchValue, 2)
			if kmpMatch || editDistanceMatch {
				filteredProgressList = append(filteredProgressList, v)
			}
		}
		result = filteredProgressList
	}

	return result, nil
}

func (d *PgDbsProgressAnime) RemoveProgress(reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary RemoveProgress: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsRelAnimeUserLibrary RemoveProgress: Rollback: %v", err)
		}
	}()

	err = DeleteProgress(tx, reqUser, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary RemoveProgress: DeleteRelAnimeUserLibrary: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary RemoveProgress: Commit: %v", err)
		return err
	}

	return nil
}

func InsertProgress(tx *sql.Tx, reqUser *User, reqAnime *Anime) (*ProgressAnime, error) {
	result := &ProgressAnime{
		Anime: &Anime{},
	}

	query := `WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	), insert_progress AS (
		INSERT INTO progress_anime (id, anime_id, user_library_id)
		SELECT $2, $3, user_lib.id
		FROM user_lib
		RETURNING id, created_at, updated_at, episode, anime_id
	)
	SELECT insert_progress.id, insert_progress.created_at, insert_progress.updated_at, insert_progress.episode,
	anime.id, anime.created_at, anime.updated_at, anime.kind, anime.episodes, anime.description, anime.image_url,
	anime_names.id, anime_names.created_at, anime_names.updated_at, anime_names.name
	FROM insert_progress
	JOIN anime ON anime.id = insert_progress.anime_id
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

func UpdateProgress(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) (*ProgressAnime, error) {
	result := &ProgressAnime{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	WITH user_lib AS (
		SELECT user_library.id FROM user_library WHERE user_id = $1
	), select_anime AS (
		SELECT anime.* FROM anime
		JOIN progress_anime ON anime.id = progress_anime.anime_id
		WHERE progress_anime.id = $2
	), update_progress AS (
		UPDATE progress_anime progress
		SET
			updated_at = $3,
			episode = $4
		FROM user_lib, select_anime
		WHERE progress.id = $2
		AND progress.user_library_id = user_lib.id
		AND $4 <= select_anime.episodes
		RETURNING progress.id, progress.created_at, progress.updated_at, progress.episode, progress.anime_id
	)
	SELECT
	update_progress.id, update_progress.created_at, update_progress.updated_at, update_progress.episode,
	anime.id, anime.created_at, anime.updated_at, anime.kind, anime.episodes, anime.description, anime.image_url,
	anime_names.id, anime_names.created_at, anime_names.updated_at, anime_names.name
	FROM update_progress
	JOIN select_anime anime ON update_progress.anime_id = anime.id
	JOIN anime_names ON anime.anime_names_id = anime_names.id
	`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Id,
		time.Now(),
		reqRelAnimeUserLibrary.Episode,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
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
		log.Printf("error: DbsRelAnimeUserLibrary UpdateRelAnimeUserLibraryProgress: Query: %v", err)
		return nil, err
	}

	return result, nil
}

func SelectProgressByAnimeId(tx *sql.Tx, reqUser *User, reqAnime *Anime) (*ProgressAnime, error) {
	result := &ProgressAnime{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	)
	SELECT
	ul.id, ul.created_at, ul.updated_at, ul.episode,
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM progress_anime ul
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

func SelectProgressById(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) (*ProgressAnime, error) {
	result := &ProgressAnime{
		Anime: &Anime{
			AlternativeNames: make([]*AnimeName, 0),
		},
	}

	query := `
	SELECT
	p.id, p.created_at, p.updated_at, p.episode,
	a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
	an.id, an.created_at, an.updated_at, an.name
	FROM progress_anime p
	JOIN anime a ON p.anime_id = a.id
	JOIN anime_names an ON a.anime_names_id = an.id
	JOIN user_library ON p.user_library_id = user_library.id
	WHERE user_library.user_id = $1
	AND p.id = $2
	`

	err := tx.QueryRow(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
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

func SelectProgressStarted(tx *sql.Tx, reqUser *User, orderBy string) ([]*ProgressAnime, error) {
	result := make([]*ProgressAnime, 0)

	query := fmt.Sprintf(`
		SELECT
		p.id, p.created_at, p.updated_at, p.episode,
		a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
		an.id, an.created_at, an.updated_at, an.name
		FROM progress_anime p
		JOIN anime a ON p.anime_id = a.id
		JOIN anime_names an ON a.anime_names_id = an.id
		JOIN user_library ON p.user_library_id = user_library.id
		WHERE user_library.user_id = $1
		AND p.episode > 0
		AND p.episode < a.episodes
		ORDER BY an.name %s
	`,
		orderBy,
	)

	queryRows, err := tx.Query(query, reqUser.Id)
	if err != nil {
		log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() {
		temp := &ProgressAnime{
			Anime: &Anime{
				AnimeName:        AnimeName{},
				AlternativeNames: make([]*AnimeName, 0),
			},
		}

		if err := queryRows.Scan(
			&temp.Id,
			&temp.CreatedAt,
			&temp.UpdatedAt,
			&temp.Episode,
			&temp.Anime.Id,
			&temp.Anime.CreatedAt,
			&temp.Anime.UpdatedAt,
			&temp.Anime.Kind,
			&temp.Anime.Episodes,
			&temp.Anime.Description,
			&temp.Anime.ImageUrl,
			&temp.Anime.AnimeName.Id,
			&temp.Anime.AnimeName.CreatedAt,
			&temp.Anime.AnimeName.UpdatedAt,
			&temp.Anime.AnimeName.Name,
		); err != nil {
			log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Scan: %v", err)
			return nil, err
		}

		result = append(result, temp)
	}

	return result, nil
}

func SelectProgressNotStarted(tx *sql.Tx, reqUser *User, orderBy string) ([]*ProgressAnime, error) {
	result := make([]*ProgressAnime, 0)

	query := fmt.Sprintf(`
		SELECT
		p.id, p.created_at, p.updated_at, p.episode,
		a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
		an.id, an.created_at, an.updated_at, an.name
		FROM progress_anime p
		JOIN anime a ON p.anime_id = a.id
		JOIN anime_names an ON a.anime_names_id = an.id
		JOIN user_library ON p.user_library_id = user_library.id
		WHERE user_library.user_id = $1
		AND p.episode = 0
		ORDER BY an.name %s
	`,
		orderBy,
	)

	queryRows, err := tx.Query(query, reqUser.Id)
	if err != nil {
		log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() {
		temp := &ProgressAnime{
			Anime: &Anime{
				AnimeName:        AnimeName{},
				AlternativeNames: make([]*AnimeName, 0),
			},
		}

		if err := queryRows.Scan(
			&temp.Id,
			&temp.CreatedAt,
			&temp.UpdatedAt,
			&temp.Episode,
			&temp.Anime.Id,
			&temp.Anime.CreatedAt,
			&temp.Anime.UpdatedAt,
			&temp.Anime.Kind,
			&temp.Anime.Episodes,
			&temp.Anime.Description,
			&temp.Anime.ImageUrl,
			&temp.Anime.AnimeName.Id,
			&temp.Anime.AnimeName.CreatedAt,
			&temp.Anime.AnimeName.UpdatedAt,
			&temp.Anime.AnimeName.Name,
		); err != nil {
			log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Scan: %v", err)
			return nil, err
		}

		result = append(result, temp)
	}

	return result, nil
}

func SelectProgressCompleted(tx *sql.Tx, reqUser *User, orderBy string) ([]*ProgressAnime, error) {
	result := make([]*ProgressAnime, 0)

	query := fmt.Sprintf(`
		SELECT
		p.id, p.created_at, p.updated_at, p.episode,
		a.id, a.created_at, a.updated_at, a.kind, a.episodes, a.description, a.image_url,
		an.id, an.created_at, an.updated_at, an.name
		FROM progress_anime p
		JOIN anime a ON p.anime_id = a.id
		JOIN anime_names an ON a.anime_names_id = an.id
		JOIN user_library ON p.user_library_id = user_library.id
		WHERE user_library.user_id = $1
		AND p.episode = a.episodes
		ORDER BY an.name %s
	`,
		orderBy,
	)

	queryRows, err := tx.Query(query, reqUser.Id)
	if err != nil {
		log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Query: %v", err)
		return nil, err
	}

	for queryRows.Next() {
		temp := &ProgressAnime{
			Anime: &Anime{
				AnimeName:        AnimeName{},
				AlternativeNames: make([]*AnimeName, 0),
			},
		}

		if err := queryRows.Scan(
			&temp.Id,
			&temp.CreatedAt,
			&temp.UpdatedAt,
			&temp.Episode,
			&temp.Anime.Id,
			&temp.Anime.CreatedAt,
			&temp.Anime.UpdatedAt,
			&temp.Anime.Kind,
			&temp.Anime.Episodes,
			&temp.Anime.Description,
			&temp.Anime.ImageUrl,
			&temp.Anime.AnimeName.Id,
			&temp.Anime.AnimeName.CreatedAt,
			&temp.Anime.AnimeName.UpdatedAt,
			&temp.Anime.AnimeName.Name,
		); err != nil {
			log.Printf("error: Dbs: RelAnimeUserLibrary: SelectProgressStarted: Scan: %v", err)
			return nil, err
		}

		result = append(result, temp)
	}

	return result, nil
}

func DeleteProgress(tx *sql.Tx, reqUser *User, reqRelAnimeUserLibrary *ProgressAnime) error {
	query := `
	WITH user_lib AS (
		SELECT * FROM user_library WHERE user_id = $1
	)
	DELETE FROM progress_anime ul
	USING user_lib
	WHERE ul.user_library_id = user_lib.id
	AND ul.id = $2
	`

	queryResult, err := tx.Exec(
		query,
		reqUser.Id,
		reqRelAnimeUserLibrary.Id,
	)
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary DeleteRelAnimeUserLibrary: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: DbsRelAnimeUserLibrary DeleteRelAnimeUserLibrary: RowsAffected: %v", err)
		return err
	}
	if n == 0 {
		log.Printf("error: DbsRelAnimeUserLibrary DeleteRelAnimeUserLibrary: RowsAffected: %v", sql.ErrNoRows)
		return sql.ErrNoRows
	}

	return nil
}
