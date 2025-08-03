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
	Delete(reqUser *User, reqJwt *tokens.Jwt) error
	DeleteExpired() error
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

func (d *PgDbsJwt) Delete(reqUser *User, reqJwt *tokens.Jwt) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: Jwt: Delete: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: Jwt: Delete: Rollback: %v", err)
		}
	}()

	err = DeleteJwt(tx, reqUser, reqJwt)
	if err != nil {
		log.Printf("error: Dbs: Jwt: Delete: DeleteJwt: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Jwt: Delete: Commit: %v", err)
		return err
	}

	return nil
}

func (d *PgDbsJwt) DeleteExpired() error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteExpired: Conn: %v", err)
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: Dbs: Jwt: DeleteExpired: Rollback: %v", err)
		}
	}()

	err = DeleteExpired(tx)
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteExpired: DeleteExpired: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteExpired: Commit: %v", err)
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

func DeleteJwt(tx *sql.Tx, reqUser *User, reqJwt *tokens.Jwt) error {
	query := `
	DELETE FROM jwt
	WHERE token = $1
	AND user_id = $2
	`

	log.Printf(reqJwt.Token.PlainText)
	reqJwtHash := tokens.HashFromPlainText(reqJwt.Token.PlainText)

	queryResult, err := tx.Exec(
		query,
		reqJwtHash[:],
		reqUser.Id,
	)
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteJwt: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteJwt: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: Jwt: DeleteJwt: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}

func DeleteExpired(tx *sql.Tx) error {
	query := `
	DELETE FROM jwt
	WHERE refresh_token_expiration < $1
	`

	queryResult, err := tx.Exec(
		query,
		time.Now(),
	)
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteExpired: Query: %v", err)
		return err
	}

	n, err := queryResult.RowsAffected()
	if err != nil {
		log.Printf("error: Dbs: Jwt: DeleteExpired: RowsAffected: %v", err)
		return err
	}

	if n == 0 {
		log.Printf("error: Dbs: Jwt: DeleteJwt: RowsAffected: %v", n)
		return sql.ErrNoRows
	}

	return nil
}
