package web

import (
	"fmt"
	"net/http"

	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
)

// getFront returns a handler for GET / requests.
func getFront(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %w", err)
			getError(s)(w, r)
			return
		}

		token, err := cSRFToken(
			"",
			s.TokenSigningSecret,
			s.CSRFTokenLifetime,
		)
		if err != nil {
			getError(s)(w, r)
			return
		}

		err = sess.Set(s.sessionValues.csrf, token)
		if err != nil {
			s.Logger.Printf("server error setting value on session because: %w", err)
			getError(s)(w, r)
			return
		}

		frontForms := Input{
			CSRF: token,
		}
		s.templates.ExecuteTemplate(w, "front.layout", frontForms)
	}
}

func postLogin(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the form data
		err := r.ParseForm()
		if err != nil {
			s.Logger.Printf("server error parsing token beccause: %w", err)
			getError(s)(w, r)
			return
		}

		loginForm := Input{
			Values:  r.PostForm,
			VErrors: ValidationErrors{},
			CSRF:    r.FormValue("_csrf"),
		}

		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %w", err)
			getError(s)(w, r)
			return
		}

		valid := validCSRF(r.FormValue("_csrf"), s.TokenSigningSecret)
		if !valid || r.FormValue("_csrf") != sess.Get(s.sessionValues.csrf) {
			s.Logger.Printf(" login attempt with incorrect CSRF token at username %s", r.FormValue("Username"))
			loginForm.VErrors.Add("generic", "Please Try Again.")

			newToken, err := cSRFToken(
				"", //no subject for login/sign up forms
				s.TokenSigningSecret,
				s.CSRFTokenLifetime,
			)
			if err != nil {
				getError(s)(w, r)
				return
			}

			err = sess.Set(s.sessionValues.csrf, newToken)
			if err != nil {
				s.Logger.Printf("server error setting value on session because: %w", err)
				getError(s)(w, r)
				return
			}
			loginForm.CSRF = newToken

			s.templates.ExecuteTemplate(w, "front.layout", loginForm)
			return
		}

		restToken, err := s.Iss1C.GetAuthToken(r.FormValue("Username"), r.FormValue("Password"))
		if err == issue1.ErrCredentialsUnaccepted {
			s.Logger.Printf("failed login attempt at username %s", r.FormValue("Username"))
			loginForm.VErrors.Add("generic", "Your username or password is wrong")
			s.templates.ExecuteTemplate(w, "front.layout", loginForm)
			return
		} else if err != nil {
			s.Logger.Printf("server error getting auth token beccause: %w", err)
			getError(s)(w, r)
			return
		}

		err = sess.Set(s.sessionValues.username, r.FormValue("Username"))
		if err != nil {
			getError(s)(w, r)
			return
		}
		err = sess.Set(s.sessionValues.restRefreshToken, restToken)
		if err != nil {
			getError(s)(w, r)
			return
		}

		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}
}

func getHome(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %w", err)
			getError(s)(w, r)
			return
		}

		w.Write([]byte(fmt.Sprintf("Welcome home %s", sess.Get(s.sessionValues.username))))
	}
}

func getError(s *Setup) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
