package database

import (
	"errors"
	"log"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/google/uuid"
)

type DbsJwt interface {
	Insert(userId uuid.UUID, ttl_token, ttl_refresh time.Duration, scope string) (*tokens.Jwt, error)
	Update(token *tokens.Jwt) (*tokens.Jwt, error)
	Get(user *User) (*tokens.Jwt, error)
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

func (d *PgDbsJwt) Update(token *tokens.Jwt) (*tokens.Jwt, error) {
	updatedToken := &tokens.Jwt{
		Token:        &tokens.Token{},
		RefreshToken: &tokens.Token{},
	}

	tx, err := d.db.Conn().Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: dbs jwt Update failed rollback: %v", err)
		}
	}()

	query := `UPDATE jwt
	SET
		updated_at = $2,
		token = $3
	WHERE id = $1
	RETURNING id, created_at, updated_at, token, refresh_token, refresh_token_expiration, scope, user_id`

	newToken, err := tokens.GenerateToken(time.Hour * 12)
	if err != nil {
		return nil, err
	}

	err = tx.QueryRow(
		query,
		token.Id,
		time.Now(),
		newToken.Hash,
	).Scan(
		&updatedToken.Id,
		&updatedToken.CreatedAt,
		&updatedToken.UpdatedAt,
		&updatedToken.Token.Hash,
		&updatedToken.RefreshToken.Hash,
		&updatedToken.RefreshToken.Expiry,
		&updatedToken.Scope,
		&updatedToken.UserId,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return updatedToken, nil
}

func (d *PgDbsJwt) Get(user *User) (*tokens.Jwt, error) {
	existingToken := tokens.Jwt{
		Token:        &tokens.Token{},
		RefreshToken: &tokens.Token{},
	}

	query := `SELECT * FROM jwt
	WHERE user_id = $1`

	err := d.db.Conn().QueryRow(
		query,
		user.Id,
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
		return nil, err
	}

	return &existingToken, nil
}

func (d *PgDbsJwt) Delete(user *User) error {
	tx, err := d.db.Conn().Begin()
	if err != nil {
		return err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
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
