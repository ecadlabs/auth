package handlers

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/ecadlabs/auth/errors"
	"github.com/ecadlabs/auth/storage"
	"github.com/ecadlabs/auth/utils"
	log "github.com/sirupsen/logrus"
)

type TokenFactory struct {
	JWTSecretGetter  func() ([]byte, error)
	JWTSigningMethod jwt.SigningMethod
	BaseURL          func() string
	Namespace        string
}

func (t *TokenFactory) Create(claims jwt.MapClaims, user *storage.User, aud string, exp time.Duration) (string, error) {
	now := time.Now()
	baseClaims := jwt.MapClaims{
		"sub": user.ID,
		"exp": now.Add(exp).Unix(),
		"iat": now.Unix(),
		"iss": t.BaseURL(),
		"aud": aud,
	}

	for i, val := range claims {
		baseClaims[utils.NSClaim(t.Namespace, i)] = val
	}

	token := jwt.NewWithClaims(t.JWTSigningMethod, baseClaims)

	secret, err := t.JWTSecretGetter()
	if err != nil {
		return "", err
	}

	return token.SignedString(secret)
}

func (t *TokenFactory) Verify(requestToken string) (*jwt.Token, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) { return t.JWTSecretGetter() })
	if err != nil {
		return nil, err
	}

	if t.JWTSigningMethod.Alg() != token.Header["alg"] {
		log.Errorf("Expected %s signing method but token specified %s", t.JWTSigningMethod.Alg(), token.Header["alg"])
		return nil, errors.ErrInvalidToken
	}

	if !token.Valid {
		log.Errorln("Invalid token")
		return nil, errors.ErrInvalidToken
	}

	return token, nil
}

func (t *TokenFactory) GetClaim(token *jwt.Token, name string) interface{} {
	claims := token.Claims.(jwt.MapClaims)

	return claims[utils.NSClaim(t.Namespace, name)]
}
