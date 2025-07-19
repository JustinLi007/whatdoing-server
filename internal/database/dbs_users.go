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
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
	Username  *string   `json:"username"`
	Email     string    `json:"email"`
	Password  Password  `json:"-"`
	Role      string    `json:"role"`
}

type DbsUsers interface {
	CreateUser(user *User) (*User, error)
	GetUserByEmailPassword(user *User) (*User, error)
	GetUserById(id uuid.UUID) (*User, error)
	AuthenticateWithJwt(jwt *tokens.Jwt) (*User, error)
}

type PgDbsUsers struct {
	db DbService
}

var AnonymousUser = &User{}
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

func (d *PgDbsUsers) CreateUser(user *User) (*User, error) {
	newUser := &User{
		Password: Password{},
	}

	tx, err := d.db.Conn().Begin()
	if err != nil {
		log.Printf("error: DbsUsers CreateUser: Conn: %v", err)
		return nil, err
	}
	defer func() {
		err := tx.Rollback()
		if err != nil {
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: DbsUsers CreateUser: Rollback: %v", err)
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

	err = InsertUserLibrary(tx, newUser)
	if err != nil {
		log.Printf("error: DbsUsers CreateUser: InsertUserLibraryStarted: %v", err)
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("error: DbsUsers CreateUser: Commit: %v", err)
		return nil, err
	}

	return newUser, nil
}

func (d *PgDbsUsers) GetUserByEmailPassword(user *User) (*User, error) {
	existingUser := &User{
		Password: Password{},
	}

	query := `SELECT * FROM users
	WHERE email = $1`

	err := d.db.Conn().QueryRow(
		query,
		user.Email,
	).Scan(
		&existingUser.Id,
		&existingUser.CreatedAt,
		&existingUser.UpdatedAt,
		&existingUser.Username,
		&existingUser.Email,
		&existingUser.Password.Hash,
		&existingUser.Role,
	)
	if err != nil {
		return nil, err
	}

	passwordMatch, err := existingUser.Password.Validate(user.Password.PlainText)
	if err != nil {
		return nil, err
	}

	if !passwordMatch {
		return nil, errors.New("error: dbsUsers GetUserByEmailPassword: failed")
	}

	return existingUser, nil
}

func (d *PgDbsUsers) GetUserById(id uuid.UUID) (*User, error) {
	user := &User{
		Password: Password{},
	}

	query := `SELECT * FROM users
	WHERE id = $1`

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

func (d *PgDbsUsers) AuthenticateWithJwt(jwt *tokens.Jwt) (*User, error) {
	existingUser := &User{
		Password: Password{},
	}

	existingJwt := &tokens.Jwt{
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
			if err.Error() == "sql: transaction has already been committed or rolled back" {
				return
			}
			log.Printf("error: dbsUsers AuthenticateByJwt Rollback: %v", err)
		}
	}()

	queryGetJwt := `SELECT * FROM jwt
	WHERE token = $1`

	hashToValidate := tokens.HashFromPlainText(jwt.Token.PlainText)

	err = tx.QueryRow(
		queryGetJwt,
		hashToValidate[:],
	).Scan(
		&existingJwt.Id,
		&existingJwt.CreatedAt,
		&existingJwt.UpdatedAt,
		&existingJwt.Token.Hash,
		&existingJwt.RefreshToken.Hash,
		&existingJwt.RefreshToken.Expiry,
		&existingJwt.Scope,
		&existingJwt.UserId,
	)
	if err != nil {
		return nil, err
	}

	if !tokens.ValidateHash(existingJwt.Token.Hash, jwt.Token.PlainText) {
		return nil, errors.New("error: dbsUsers AuthenticateByJwt: failed")
	}

	queryGetUser := `SELECT * FROM users
	WHERE id = $1`

	err = tx.QueryRow(
		queryGetUser,
		existingJwt.UserId,
	).Scan(
		&existingUser.Id,
		&existingUser.CreatedAt,
		&existingUser.UpdatedAt,
		&existingUser.Username,
		&existingUser.Email,
		&existingUser.Password.Hash,
		&existingUser.Role,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return existingUser, nil
}
