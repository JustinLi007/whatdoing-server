package database

import (
	"errors"
	"log"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Password struct {
	PlainText string
	Hash      []byte
}

func (p *Password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	p.PlainText = plainTextPassword
	p.Hash = hash

	return nil
}

func (p *Password) Validate(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.Hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

type User struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Username  *string   `json:"username"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Role      string    `json:"role"`
}

type DbsUsers interface {
	CreateUser(user User) (*User, error)
	GetUserById(id uuid.UUID) (*User, error)
	AuthenticateByJwt(jwt *tokens.Jwt) (*User, error)
}

type PgDbsUsers struct {
	db DbService
}

var dbsUsersInstance *PgDbsUsers

func NewDbsUsers(db DbService) DbsUsers {
	if dbsUsersInstance != nil {
		return dbsUsersInstance
	}

	newDbsUsers := &PgDbsUsers{
		db: db,
	}
	dbsUsersInstance = newDbsUsers

	return dbsUsersInstance
}

func (d *PgDbsUsers) CreateUser(user User) (*User, error) {
	log.Printf("dbsUsers CreateUser reached. Not implemented.")

	newUser := &User{
		Password: Password{},
	}
	tx, err := d.db.Conn().Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("error: dbsUsers CreateUser Rollback: %v", err)
		}
	}()

	query := `INSERT INTO users (id, email, password_hash)
	VALUES ($1, $2, $3)
	RETURNING id, created_at, updated_at, username, email, password_hash, role;`

	err = tx.QueryRow(
		query,
		uuid.New(),
		user.Email,
		user.Password.Hash,
	).Scan(
		&newUser.Id,
		&newUser.CreatedAt,
		&newUser.UpdatedAt,
		&newUser.Username,
		&newUser.Email,
		&newUser.Password.Hash,
		&newUser.Role,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (d *PgDbsUsers) GetUserById(id uuid.UUID) (*User, error) {
	user := &User{}

	query := `SELECT FROM users
	WHERE id = $1
	RETURNING id, created_at, updated_at, username, email, password_hash, role;`

	err := d.db.Conn().QueryRow(
		query,
		id,
	).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (d *PgDbsUsers) AuthenticateByJwt(jwt *tokens.Jwt) (*User, error) {
	user := &User{
		Password: Password{},
	}

	dbJwt := tokens.Jwt{
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
			log.Printf("error: dbsUsers AuthenticateByJwt Rollback: %v", err)
		}
	}()

	queryGetJwt := `SELECT * FROM jwt
	WHERE user_id = $1
	RETURNING id, created_at, updated_at, token, refresh_token, refresh_token_expiration, scope, user_id;`

	err = tx.QueryRow(
		queryGetJwt,
		jwt.UserId,
	).Scan(
		&dbJwt.Id,
		&dbJwt.CreatedAt,
		&dbJwt.UpdatedAt,
		&dbJwt.Token.Hash,
		&dbJwt.RefreshToken.Hash,
		&dbJwt.RefreshToken.Expiry,
		&dbJwt.Scope,
		&dbJwt.UserId,
	)
	if err != nil {
		return nil, err
	}

	if !tokens.ValidateHash(dbJwt.Token.Hash, jwt.Token.PlainText) {
		return nil, errors.New("error: dbsUsers AuthenticateByJwt: failed")
	}

	queryGetUser := `SELECT * FROM users
	WHERE id = $1
	RETURNING id, created_at, updated_at, username, email, password_hash, role;`

	err = tx.QueryRow(
		queryGetUser,
		jwt.UserId,
	).Scan(
		&user.Id,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Username,
		&user.Email,
		&user.Password.Hash,
		&user.Role,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return user, nil
}
