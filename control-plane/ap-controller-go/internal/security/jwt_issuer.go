package security

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret []byte, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{
		secret: secret,
		ttl:    ttl,
	}
}

func (i *JWTIssuer) Issue(ctx context.Context, subject string) (string, int64, error) {
	now := time.Now()
	exp := now.Add(i.ttl)

	claims := jwt.MapClaims{
		"sub": subject,
		"iat": now.Unix(),
		"exp": exp.Unix(),
		"iss": "ap-controller",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(i.secret)
	if err != nil {
		return "", 0, err
	}

	return s, int64(i.ttl.Seconds()), nil
}
