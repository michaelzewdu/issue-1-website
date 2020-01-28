package web

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
	"net/url"
	"time"

	"github.com/slim-crown/issue-1-website/internal/services/session"
)

type sessionValues struct {
	restRefreshToken string
	username         string
	csrf             string
}

// SessionTokenClaims specifies custom JWT claim used for sessions.
type SessionTokenClaims struct {
	jwt.StandardClaims
	SessionID       string `json:"sessionID,omitempty"`
	RestAccessToken string `json:"token,omitempty"`
}

// sessionStart looks for a sessionID on the request cookies and returns the
// session under it if found. If not found, it creates a new session and attaches
// a new cookie.
func sessionStart(s *Setup, w http.ResponseWriter, r *http.Request) (*session.Session, error) {
	cookie, err := r.Cookie(s.CookieName)
	if err == nil && cookie.Value != "" {
		// if session found on cookie
		sid, _ := url.QueryUnescape(cookie.Value)
		sess, errs := s.SessionService.GetSession(sid)
		// TODO session not found
		sessionFound := true
		for _, err := range errs {
			if gorm.IsRecordNotFoundError(err) {
				sessionFound = false
			} else {
				return nil, fmt.Errorf("unable to retrieve session because: %+v", errs)
			}
		}
		if sessionFound {
			// TODO check if max age gets updated
			//cookie.MaxAge = int(s.SessionHardLifetime.Seconds())
			w.Header().Set("Set-Cookie", cookie.String())
			return sess, nil
		}
	}
	sess, errs := s.SessionService.NewSession(generateRandomID(32), s.SessionHardLifetime)

	if len(errs) > 0 {
		return nil, fmt.Errorf("unable to create session because: %+v", errs)
	}

	cookie = &http.Cookie{
		Name:     s.CookieName,
		Value:    sess.UUID,
		MaxAge:   int(s.SessionHardLifetime.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Secure:   s.HTTPS,
		HttpOnly: true,
	}

	w.Header().Set("Set-Cookie", cookie.String())

	//fmt.Printf("Session: %+v\nCookie:%+v\n", sess, cookie)
	return sess, nil
}

var errNotLoggedIn = errors.New("session: session found not logged in")

var errRefreshTokenExpired = errors.New("session: refresh token found on session is expired")

// SessionStartLoggedIn is a wrapper around sessionStart that assures the returned session is a logged
// in one. It also redirects to the login Page if not so one can simply return after using it.
// If startSession returns error, it'll also display the error Page and one can simply return as well.
func SessionStartLoggedIn(s *Setup, w http.ResponseWriter, r *http.Request) (*session.Session, error) {
	sess, err := sessionStart(s, w, r)
	if err != nil {
		s.Logger.Printf("server error starting session because: %v", err)
		showErrorPage(w, r)
		return nil, err
	}
	if sess.Get(s.sessionValues.username) == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil, errNotLoggedIn
	}
	return sess, nil
}

func refreshTokenAuthOnSession(sess *session.Session, s *Setup, w http.ResponseWriter, r *http.Request) error {
	authToken := sess.Get(s.sessionValues.restRefreshToken)
	authToken, err := s.Iss1C.RefreshAuthToken(authToken)
	switch err {
	case nil:
		err = sess.Set(s.sessionValues.restRefreshToken, authToken)
		if err != nil {
			s.Logger.Printf("server error setting auth token on session because: %v", err)
			showErrorPage(w, r)
			return err
		}
		return nil
	case issue1.ErrAccessDenied:
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return errRefreshTokenExpired
	default:
		s.Logger.Printf("server error refreshing token refreshing token because: %v", err)
		showErrorPage(w, r)
		return err
	}
}

// sessionDestroy removes all cookies set by session start.
func sessionDestroy(s *Setup, w http.ResponseWriter, r *http.Request) error {
	// TODO test
	cookie, err := r.Cookie(s.CookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	_, errs := s.SessionService.DeleteSession(cookie.Value)
	if len(errs) > 0 {
		return fmt.Errorf("unable to destroy session because: %+v", errs)
	}
	cookie = &http.Cookie{
		Name:     s.CookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}
	w.Header().Set("Set-Cookie", cookie.String())

	return nil
}
