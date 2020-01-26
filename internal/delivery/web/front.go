package web

import (
	"fmt"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
)

// getFront returns a handler for GET / requests.
func getFront(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %v", err)
			getError(s)(w, r)
			return
		}
		if sess.Get(s.sessionValues.username) != "" {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
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
			s.Logger.Printf("server error setting value on session because: %v", err)
			getError(s)(w, r)
			return
		}

		frontForms := Input{
			CSRF: token,
		}
		_ = s.templates.ExecuteTemplate(w, "front.layout", frontForms)
	}
}

func postLogin(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse the form data
		loginForm := Input{
			VErrors: ValidationErrors{},
		}
		err := r.ParseForm()
		if err != nil {
			s.Logger.Printf("server error parsing token beccause: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		loginForm.Values = r.PostForm
		loginForm.CSRF = r.FormValue("_csrf")

		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %v", err)
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
				s.Logger.Printf("server error setting value on session because: %v", err)
				getError(s)(w, r)
				return
			}
			loginForm.CSRF = newToken

			w.WriteHeader(http.StatusBadRequest)
			_ = s.templates.ExecuteTemplate(w, "login.form", loginForm)
			return
		}

		restToken, err := s.Iss1C.GetAuthToken(r.FormValue("Username"), r.FormValue("Password"))
		switch err {
		case nil:
			// restart the session
			err = sessionDestroy(s, w, r)
			if err != nil {
				getError(s)(w, r)
				return
			}
			sess, err = sessionStart(s, w, r)
			if err != nil {
				s.Logger.Printf("server error starting session because: %v", err)
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
		case issue1.ErrCredentialsUnaccepted:
			s.Logger.Printf("failed login attempt at username %s", r.FormValue("Username"))
			loginForm.VErrors.Add("generic", "Your username or password is wrong")
			w.WriteHeader(http.StatusUnauthorized)
			_ = s.templates.ExecuteTemplate(w, "login.form", loginForm)
		default:
			s.Logger.Printf("server error getting auth token beccause: %v", err)
			loginForm.VErrors.Add("generic", "Server Error. Please Try Again Later.")
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.templates.ExecuteTemplate(w, "login.form", loginForm)
		}
	}
}

func postSignUp(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signUpForm := Input{
			VErrors: ValidationErrors{},
		}
		err := r.ParseForm()
		if err != nil {
			s.Logger.Printf("server error parsing token beccause: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		signUpForm.Values = r.PostForm
		signUpForm.CSRF = r.FormValue("_csrf")

		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %v", err)
			getError(s)(w, r)
			return
		}

		valid := validCSRF(r.FormValue("_csrf"), s.TokenSigningSecret)
		if !valid || r.FormValue("_csrf") != sess.Get(s.sessionValues.csrf) {
			s.Logger.Printf(" login attempt with incorrect CSRF token at username %s", r.FormValue("Username"))
			signUpForm.VErrors.Add("generic", "Please Try Again.")

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
				s.Logger.Printf("server error setting value on session because: %v", err)
				getError(s)(w, r)
				return
			}
			signUpForm.CSRF = newToken

			w.WriteHeader(http.StatusBadRequest)
			_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
			return
		}

		// Validate the form contents
		signUpForm.Required("FirstName", "Username", "Email", "Password", "PasswordConfirm")
		signUpForm.MatchesPattern("Email", EmailRX)
		// todo username regex
		signUpForm.MinLength("Password", 8)
		signUpForm.MinLength("Username", 5)
		signUpForm.MaxLength("Username", 24)
		signUpForm.PasswordMatches("Password", "PasswordConfirm")

		// If there are any errors, redisplay the signup form.
		if !signUpForm.Valid() {
			w.WriteHeader(http.StatusBadRequest)
			_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
			return
		}

		user := &issue1.User{
			Username:   r.FormValue("Username"),
			Email:      r.FormValue("Email"),
			FirstName:  r.FormValue("FirstName"),
			MiddleName: r.FormValue("MiddleName"),
			LastName:   r.FormValue("LastName"),
			Password:   r.FormValue("Password"),
		}
		user, err = s.Iss1C.UserService.AddUser(user)
		switch err {
		case nil:
			// if account creation successful, log user in.
			restToken, err := s.Iss1C.GetAuthToken(user.Username, r.FormValue("Password"))
			switch err {
			case nil:
				// restart the session
				err = sessionDestroy(s, w, r)
				if err != nil {
					getError(s)(w, r)
					return
				}
				sess, err = sessionStart(s, w, r)
				if err != nil {
					s.Logger.Printf("server error starting session because: %v", err)
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
			case issue1.ErrCredentialsUnaccepted:
				s.Logger.Printf("failed login attempt at username %s", r.FormValue("Username"))
				signUpForm.VErrors.Add("generic", "Success. Try logging in.")
				w.WriteHeader(http.StatusUnauthorized)
				_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
			default:
				s.Logger.Printf("server error getting auth token beccause: %v", err)
				//getError(s)(w, r)
				signUpForm.VErrors.Add("generic", "Success. Try logging in.")
				w.WriteHeader(http.StatusInternalServerError)
				_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
			}
		case issue1.ErrUserNameOccupied:
			s.Logger.Printf("signup attempt on a occupied username")
			//getError(s)(w, r)
			signUpForm.VErrors.Add("Username", "Username is occupied.")
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
		case issue1.ErrEmailIsOccupied:
			s.Logger.Printf("signup attempt on a occupied email")
			signUpForm.VErrors.Add("Email", "Email is occupied.")
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.templates.ExecuteTemplate(w, "signup.form", signUpForm)
		default:
			s.Logger.Printf("server error getting auth token beccause: %v", err)
			signUpForm.VErrors.Add("generic", "Server Error. Please Try Again Later.")
			w.WriteHeader(http.StatusInternalServerError)
			_ = s.templates.ExecuteTemplate(w, "login.form", signUpForm)
		}
	}
}

func getHome(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := sessionStart(s, w, r)
		if err != nil {
			s.Logger.Printf("server error starting session because: %v", err)
			getError(s)(w, r)
			return
		}

		w.Write([]byte(fmt.Sprintf("Welcome home %s", sess.Get(s.sessionValues.username))))
	}
}

func getError(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
