package web

import (
	"github.com/slim-crown/issue-1-website/internal/services/session"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
	"strconv"
	"time"
)

func getPostView(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := getParametersFromRequestAsMap(r)
		postIDRaw := vars["postID"]
		postID, err := strconv.Atoi(postIDRaw)
		if err != nil || postID < 1 {
			show404Page(w, r)
			return
		}
		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			return
		}
		var postData struct {
			*issue1.Post
			Releases []*issue1.Release
			*NavBarData
		}
		postData.NavBarData, err = getNavbarData(s, sess, w, r)
		if err != nil {
			return
		}
		postData.Post, err = s.Iss1C.PostService.GetPost(uint(postID))
		if err != nil {
			if err == issue1.ErrPostNotFound {
				show404Page(w, r)
				return
			}
			showErrorPage(w, r)
			return
		}
		postData.Releases = make([]*issue1.Release, 0)
		for _, id := range postData.Post.ContentsID {
			rel, err := s.Iss1C.ReleaseService.GetRelease(id)
			if err != nil {
				showErrorPage(w, r)
				return
			}
			postData.Releases = append(postData.Releases, rel)
		}
		_ = s.templates.ExecuteTemplate(w, "post.view", postData)
	}
}

type NavBarData struct {
	Username string
	Subs     map[time.Time]*issue1.Channel
}

func getNavbarData(s *Setup, sess *session.Session, w http.ResponseWriter, r *http.Request) (*NavBarData, error) {
	navData := NavBarData{}
	username := sess.Get(s.sessionValues.username)
	authToken := sess.Get(s.sessionValues.restRefreshToken)
	navData.Username = username
	subs, err := s.Iss1C.FeedService.GetFeedSubscriptions(username, authToken, issue1.SortBySubscriptionTime, issue1.SortDescending)
	if err != nil {
		if err == issue1.ErrAccessDenied {
			err = refreshTokenAuthOnSession(sess, s, w, r)
			if err != nil {
				return nil, err
			}
			subs, err = s.Iss1C.FeedService.GetFeedSubscriptions(username, sess.Get(s.sessionValues.restRefreshToken),
				issue1.SortBySubscriptionTime, issue1.SortDescending)
			if err != nil {
				showErrorPage(w, r)
				return nil, err
			}
		} else {
			showErrorPage(w, r)
			return nil, err
		}
	}
	navData.Subs = subs
	return &navData, nil
}
