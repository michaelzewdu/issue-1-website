package web

import (
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
	"time"
)

func getAccountView(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Printf("fkdls")
		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			s.Logger.Printf("%s", err.Error())
			return
		}

		username := sess.Get(s.sessionValues.username)
		authToken := sess.Get(s.sessionValues.restRefreshToken)
		var UserAccountData struct {
			User            *issue1.User
			BookmarkedPosts map[time.Time]*issue1.Post
			*NavBarData
		}
		UserAccountData.NavBarData, err = getNavbarData(s, sess, w, r)
		if err != nil {
			s.Logger.Printf("%s", err.Error())
			return
		}
		UserAccountData.User, err = s.Iss1C.UserService.GetUser(username)
		if err != nil {
			s.Logger.Printf("%s", err.Error())
			return
		}

		UserAccountData.BookmarkedPosts, err = s.Iss1C.UserService.GetUserBookmarks(username, authToken)
		if err != nil {
			s.Logger.Printf("%s", err.Error())
			return
		}

		_ = s.templates.ExecuteTemplate(w, "account", UserAccountData)

	}
}
