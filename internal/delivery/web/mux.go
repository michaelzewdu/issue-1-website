package web

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	mrand "math/rand"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/slim-crown/issue-1-website/internal/services/session"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"

	"github.com/julienschmidt/httprouter"
)

// Setup is used to inject dependencies and other required data used by the handlers.
type Setup struct {
	Config
	Dependencies
	templates *template.Template
}

// ParseTemplates is used to refresh the templates from disk.
func (s *Setup) ParseTemplates() error {
	temp, err := template.ParseGlob(s.TemplatesStoragePath + "/*")
	fmt.Printf("%s\n", temp.DefinedTemplates())
	if err != nil {
		return err
	}
	s.templates = temp
	return nil
}

// Dependencies contains dependencies used by the handlers.
type Dependencies struct {
	Iss1C          *issue1.Client
	Logger         *log.Logger
	sessionValues  sessionValues
	SessionService session.Service
}

// Config contains the different settings used to set up the handlers
type Config struct {
	TemplatesStoragePath, AssetStoragePath, AssetServingRoute string
	HostAddress, Port                                         string
	CookieName                                                string
	CSRFTokenLifetime                                         time.Duration
	SessionIdleLifetime, SessionHardLifetime                  time.Duration
	TokenSigningSecret                                        []byte
	HTTPS                                                     bool
}
type sessionValues struct {
	restRefreshToken string
	username         string
	csrf             string
}

// NewMux returns a fully configured issue1 website server.
func NewMux(s *Setup) *httprouter.Router {
	mainRouter := httprouter.New()

	err := s.ParseTemplates()
	if err != nil {
		s.Logger.Fatalf("error: initial template parsing failed because: %w\n fatal: server start-up aborted.", err)
	}

	s.sessionValues.restRefreshToken = "restRefreshToken"
	s.sessionValues.csrf = "CSRF"
	s.sessionValues.username = "username"

	fs := http.FileServer(http.Dir(s.AssetStoragePath))
	mainRouter.Handler("GET", s.AssetServingRoute+"*filepath", http.StripPrefix(s.AssetServingRoute, fs))

	mainRouter.HandlerFunc("GET", "/", getFront(s))
	mainRouter.HandlerFunc("POST", "/login", postLogin(s))
	mainRouter.HandlerFunc("GET", "/home", getHome(s))

	return mainRouter
}

// SessionTokenClaims specifies custom JWT claim used for sessions.
type SessionTokenClaims struct {
	jwt.StandardClaims
	SessionID       string `json:"sessionID"`
	RestAccessToken string `json:"token"`
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
		Subject:   tokenSubject, //no subject for login/signup forms
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
