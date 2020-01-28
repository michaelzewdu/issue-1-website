package web

import (
	"encoding/json"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
)

func getHome(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			return
		}
		var homeData struct {
			*NavBarData
		}
		homeData.NavBarData, err = getNavbarData(s, sess, w, r)
		if err != nil {
			return
		}
		_ = s.templates.ExecuteTemplate(w, "home", homeData)
	}
}

func postFeedPosts(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			return
		}
		var p struct {
			Page    uint               `json:"page"`
			PerPage uint               `json:"perPage"`
			Sorting issue1.FeedSorting `json:"sorting"`
		}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "invalid request JSON",
				http.StatusBadRequest)
			return
		}

		type augmentedPost struct {
			*issue1.Post
			Releases []*issue1.Release
		}
		postList := make([]augmentedPost, 0)
		username := sess.Get(s.sessionValues.username)
		authToken := sess.Get(s.sessionValues.restRefreshToken)
		posts, err := s.Iss1C.FeedService.GetFeedPostsPaged(p.Page, p.PerPage, p.Sorting, username, authToken)
		if err != nil {
			if err == issue1.ErrAccessDenied {
				err = refreshTokenAuthOnSession(sess, s, w, r)
				if err != nil {
					return
				}
				posts, err = s.Iss1C.FeedService.GetFeedPostsPaged(p.Page, p.PerPage, p.Sorting, username, sess.Get(s.sessionValues.restRefreshToken))
				if err != nil {
					showErrorPage(w, r)
					return
				}
			} else {
				showErrorPage(w, r)
				return
			}
		}

		for _, p := range posts {
			//releases, err := s.Iss1C.PostService.GetPostReleases(p.ID)
			//if err != nil {
			//	showErrorPage(w, r)
			//	return
			//}
			releases := make([]*issue1.Release, 0)
			for _, id := range p.ContentsID {
				rel, err := s.Iss1C.ReleaseService.GetRelease(id)
				if err != nil {
					showErrorPage(w, r)
					return
				}
				releases = append(releases, rel)
			}
			postList = append(postList, augmentedPost{
				Post:     p,
				Releases: releases,
			})
		}

		_ = s.templates.ExecuteTemplate(w, "post.list", postList)
	}
}
