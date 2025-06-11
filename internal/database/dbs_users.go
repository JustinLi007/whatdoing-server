package database

import (
	"errors"
	"log"
	"time"

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
}

type PgDbsUsers struct {
	db DbService
}

var dbsInstance *PgDbsUsers

func NewDbsUsers(db DbService) DbsUsers {
	if dbsInstance != nil {
		return dbsInstance
	}

	newDbsUsers := &PgDbsUsers{
		db: db,
	}
	dbsInstance = newDbsUsers

	return dbsInstance
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

	query := `INSERT INTO users id, email, password_hash
	VALUE ($1, $2, $3)
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
