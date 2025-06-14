package tokens

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"time"

	"github.com/google/uuid"
)

const (
	ScopeAuthenticate = "authentication"
)

type Jwt struct {
	Id           uuid.UUID `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
	Token        *Token    `json:"token"`
	RefreshToken *Token    `json:"refresh_token"`
	UserId       uuid.UUID `json:"-"`
	Scope        string    `json:"-"`
}

type Token struct {
	PlainText string    `json:"plain_text"`
	Hash      []byte    `json:"-"`
	Expiry    time.Time `json:"expiry"`
}

func GenerateJwt(userId uuid.UUID, ttl_token, ttl_refresh time.Duration, scope string) (*Jwt, error) {
	jwt := &Jwt{
		UserId: userId,
		Scope:  scope,
	}

	token, err := GenerateToken(ttl_token)
	if err != nil {
		return nil, err
	}
	refreshToken, err := GenerateToken(ttl_refresh)
	if err != nil {
		return nil, err
	}

	jwt.Token = token
	jwt.RefreshToken = refreshToken

	return jwt, nil
}

func GenerateToken(ttl time.Duration) (*Token, error) {
	token := &Token{
		Expiry: time.Now().Add(ttl),
	}

	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)
	hash := sha256.Sum256([]byte(token.PlainText))
	token.Hash = hash[:]

	return token, nil
}

func ValidateHash(secret []byte, plainText string) bool {
	hash := sha256.Sum256([]byte(plainText))
	if len(secret) == len(hash) && subtle.ConstantTimeCompare(secret, hash[:]) == 1 {
		return true
	}
	return false
}
