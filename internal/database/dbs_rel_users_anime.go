package database

import (
	"database/sql"
	"log"
	"time"

	"github.com/google/uuid"
)

type RelUsersAnime struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserId    uuid.UUID `json:"user_id"`
	AnimeId   uuid.UUID `json:"anime_id"`
}

type DbsRelUsersAnime interface {
	InsertRel(rel *RelUsersAnime) error
	GetRel(rel *RelUsersAnime) (*RelUsersAnime, error)
}

type PgDbsUsersAnime struct {
	db DbService
}

var dbsUsersAnimeInstance *PgDbsUsersAnime

func NewDbsUsersAnime(db DbService) DbsRelUsersAnime {
	if dbsUsersAnimeInstance != nil {
		return dbsUsersAnimeInstance
	}

	newDbsUsersAnime := PgDbsUsersAnime{
		db: db,
	}
	dbsUsersAnimeInstance = &newDbsUsersAnime

	return dbsUsersAnimeInstance
}

func (d *PgDbsUsersAnime) InsertRel(rel *RelUsersAnime) error {
	query := `INSERT INTO rel_users_anime (id, user_id, anime_id)
	VALUES ($1, $2, $3)`

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: dbs rel_users_anime InsertRel: Rollback: %v", err)
		}
	}()

	result, err := tx.Exec(
		query,
		uuid.New(),
		rel.UserId,
		rel.AnimeId,
	)
	if err != nil {
		return err
	}

	// FIX: refactor every RowsAffected check
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *PgDbsUsersAnime) GetRel(rel *RelUsersAnime) (*RelUsersAnime, error) {
	existingRel := &RelUsersAnime{}

	query := `SELECT * FROM rel_users_anime
	WHERE user_id = $1
	AND anime_id = $2`

	err := d.db.Conn().QueryRow(
		query,
		rel.UserId,
		rel.AnimeId,
	).Scan(
		&existingRel.Id,
		&existingRel.CreatedAt,
		&existingRel.UpdatedAt,
		&existingRel.UserId,
		&existingRel.AnimeId,
	)
	if err != nil {
		return nil, err
	}

	return existingRel, nil
}
