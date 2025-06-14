package database

import (
	"errors"
	"log"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/google/uuid"
)

type DbsJwt interface {
	CreateToken(token *tokens.Jwt) error
	UpdateToken(token *tokens.Jwt) error
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

func (d *PgDbsJwt) CreateToken(token *tokens.Jwt) error {
	query := `INSERT INTO jwt (id, token, refresh_token, refresh_token_expiration, scope, user_id)
	VALUES ($1, $2, $3, $4, $5, $6);`

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
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
		return err
	}

	n, err := result.RowsAffected()
	if err == nil {
		if n == 0 {
			return errors.New("error: dbs jwt CreateToken, failed to insert token.")
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *PgDbsJwt) UpdateToken(token *tokens.Jwt) error {
	existingToken := &tokens.Jwt{
		Token:        &tokens.Token{},
		RefreshToken: &tokens.Token{},
	}

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: dbs jwt UpdateToken failed rollback: %v", err)
		}
	}()

	queryGetToken := `SELECT * FROM token
	WHERE user_id = $1
	AND hash = $2
	RETURNING id, created_at, updated_at, token, refresh_token, refresh_token_expiration, scope, user_id;`

	err = tx.QueryRow(
		queryGetToken,
		token.Token.Hash,
		token.UserId,
	).Scan(
		&existingToken.Id,
		&existingToken.CreatedAt,
		&existingToken.UpdatedAt,
		&existingToken.Token.Hash,
		&existingToken.RefreshToken.Hash,
		&existingToken.RefreshToken.Expiry,
		&existingToken.Scope,
		&existingToken.UserId,
	)
	if err != nil {
		return err
	}

	if !tokens.ValidateHash(existingToken.Token.Hash, token.Token.PlainText) {
		return errors.New("error: dbsUsers AuthenticateByJwt: failed")
	}

	newToken, err := tokens.GenerateToken(time.Hour * 24)
	if err != nil {
		return err
	}
	newRefreshToken, err := tokens.GenerateToken(time.Hour * 24)
	if err != nil {
		return err
	}

	existingToken.UpdatedAt = time.Now()
	existingToken.Token = newToken
	existingToken.RefreshToken = newRefreshToken

	queryUpdateToken := `UPDATE jwt
	SET
		updated_at = $2,
		token = $3,
		refresh_token = $4,
		refresh_token_expiration = $5
	WHERE id = $1;`

	result, err := tx.Exec(
		queryUpdateToken,
		existingToken.Id,
		existingToken.Token.Hash,
		existingToken.RefreshToken.Hash,
		existingToken.RefreshToken.Expiry,
	)
	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("")
	}

	err = tx.Commit()
	if err == nil {
		if n == 0 {
			return errors.New("error: dbs jwt UpdateToken, failed to update token.")
		}
	}

	return nil
}
