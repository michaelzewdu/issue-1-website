package web

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	mrand "math/rand"
)

func getParametersFromRequestAsMap(r *http.Request) map[string]string {
	params := httprouter.ParamsFromContext(r.Context())
	vars := make(map[string]string, 0)
	for _, param := range params {
		vars[param.Key] = param.Value
	}
	return vars
}

// generateJWT generates jwt token from the given claims.
func generateJWT(signingKey []byte, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString(signingKey)
	return signedString, err
}

// valid validates a given token
func valid(signedToken string, signingKey []byte) (bool, error) {
	token, err := jwt.ParseWithClaims(signedToken, &SessionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})

	if err != nil {
		return false, err
	}

	if _, ok := token.Claims.(*SessionTokenClaims); !ok || !token.Valid {
		return false, err
	}

	return true, nil
}

// CSRFToken Generates random string for CSRF
func cSRFToken(tokenSubject string, signingKey []byte, lifetime time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	token.Claims = jwt.StandardClaims{
		ExpiresAt: time.Now().Add(lifetime).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   tokenSubject, //no subject for login/sign up forms
	}
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("token signing failed because %v", err)
	}
	return tokenString, nil
}

// ValidCSRF checks if a given csrf token is valid
func validCSRF(signedToken string, signingKey []byte) bool {
	if signedToken == "" {
		return false
	}
	token, _ := jwt.Parse(signedToken,
		func(token *jwt.Token) (interface{}, error) {
			return signingKey, nil
		})
	if token == nil || !token.Valid {
		return false
	}

	return true
}

func showErrorPage(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func show404Page(w http.ResponseWriter, r *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}

// GenerateRandomBytes returns securely generated random bytes.
func generateRandomBytes(n int) ([]byte, error) {
	mrand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a URL-safe, base64 encoded securely generated random string.
func generateRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateRandomID generates random id for a session
func generateRandomID(s int) string {
	mrand.Seed(time.Now().UnixNano())

	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, s)
	for i := range b {
		b[i] = letterBytes[mrand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
