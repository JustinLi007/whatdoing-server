package database

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/google/uuid"
)

type DbsJwt interface {
	Insert(userId uuid.UUID, ttl_token, ttl_refresh time.Duration, scope string) (*tokens.Jwt, error)
	Get(reqUser *User) (*tokens.Jwt, error)
	Delete(user *User) error
}

type PgDbsJwt struct {
	db DbService
}

var dbsJwtInstance *PgDbsJwt

func NewDbsJwt(db DbService) DbsJwt {
	if dbsJwtInstance != nil {
		return dbsJwtInstance
	}

	newDbsJwt := &PgDbsJwt{
		db: db,
	}
	dbsJwtInstance = newDbsJwt

	return dbsJwtInstance
}

func (d *PgDbsJwt) Insert(userId uuid.UUID, ttl_token, ttl_refresh time.Duration, scope string) (*tokens.Jwt, error) {
	token, err := tokens.GenerateJwt(userId, ttl_token, ttl_refresh, scope)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO jwt (id, token, refresh_token, refresh_token_expiration, scope, user_id)
	VALUES ($1, $2, $3, $4, $5, $6);`

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
			log.Printf("error: dbs jwt CreateToken failed rollback: %v", err)
		}
	}()

	result, err := tx.Exec(
		query,
		uuid.New(),
		token.Token.Hash,
		token.RefreshToken.Hash,
		token.RefreshToken.Expiry,
		token.Scope,
		token.UserId,
	)
	if err != nil {
		return nil, err
	}

	n, err := result.RowsAffected()
	if err == nil {
		if n == 0 {
			return nil, errors.New("error: dbs jwt CreateToken, failed to insert token.")
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (d *PgDbsJwt) Get(reqUser *User) (*tokens.Jwt, error) {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: Jwt: Get: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: Jwt: Get: Rollback: %v", err)
		}
	}()

	dbJwt, err := SelectJwtByUserId(tx, reqUser)
	if err != nil {
		log.Printf("error: Dbs: Jwt: Get: SelectJwt: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Jwt: Get: Commit: %v", err)
		return nil, err
	}

	return dbJwt, nil
}

func (d *PgDbsJwt) Delete(user *User) error {
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
			log.Printf("error: dbs jwt Delete failed rollback: %v", err)
		}
	}()

	query := `DELETE FROM jwt
	WHERE user_id = $1`

	result, err := tx.Exec(
		query,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err == nil {
		if n == 0 {
			return errors.New("error: dbs jwt Delete, failed to delete token(s).")
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func SelectJwtByUserId(tx *sql.Tx, reqUser *User) (*tokens.Jwt, error) {
	result := &tokens.Jwt{
		Token:        &tokens.Token{},
		RefreshToken: &tokens.Token{},
	}

	query := `SELECT * FROM jwt
	WHERE user_id = $1
	ORDER BY created_at DESC
	LIMIT 1`

	err := tx.QueryRow(
		query,
		reqUser.Id,
	).Scan(
		&result.Id,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.Token.Hash,
		&result.RefreshToken.Hash,
		&result.RefreshToken.Expiry,
		&result.Scope,
		&result.UserId,
	)
	if err != nil {
		log.Printf("error: Dbs: Jwt: SelectJwt: Scan: %v", err)
		return nil, err
	}

	return result, nil
}
