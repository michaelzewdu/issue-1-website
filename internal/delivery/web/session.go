package web

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/slim-crown/issue-1-website/internal/services/session"
)

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
			http.SetCookie(w, cookie)
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

	http.SetCookie(w, cookie)
	fmt.Printf("Session: %+v\nCookie:%+v\n", sess, cookie)
	return sess, nil
}

// sessionStart looks for a sessionID on the request cookies and returns the
// session under it if found. If not found, it creates a new session and attaches
// a new cookie.
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
	http.SetCookie(w, cookie)
	return nil
}
