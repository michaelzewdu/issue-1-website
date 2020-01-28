package web

import (
	"encoding/json"
	"github.com/slim-crown/issue-1-website/internal/services/session"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
	"net/http"
	"strconv"
	"time"
)

func postComment(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := getParametersFromRequestAsMap(r)
		postIDRaw := vars["postID"]
		postID, err := strconv.Atoi(postIDRaw)
		if err != nil || postID < 1 {
			show404Page(w, r)
			return
		}
		var temp struct {
			Comment string
			CSRF    string
		}
		err = json.NewDecoder(r.Body).Decode(&temp)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "invalid request JSON",
				http.StatusBadRequest)
			return
		}

		sess, err := SessionStartLoggedIn(s, w, r)
		if err != nil {
			return
		}

		valid := validCSRF(temp.CSRF, s.TokenSigningSecret)
		if !valid || temp.CSRF != sess.Get(s.sessionValues.csrf) {
			s.Logger.Printf(" login attempt with incorrect CSRF token at username %s", r.FormValue("Username"))
			newToken, err := cSRFToken(
				"", //no subject for login/sign up forms
				s.TokenSigningSecret,
				s.CSRFTokenLifetime,
			)
			if err != nil {
				showErrorPage(w, r)
				return
			}

			err = sess.Set(s.sessionValues.csrf, newToken)
			if err != nil {
				s.Logger.Printf("server error setting value on session because: %v", err)
				showErrorPage(w, r)
				return
			}
			commentForm := Input{
				CSRF: newToken,
			}

			w.WriteHeader(http.StatusBadRequest)
			_ = s.templates.ExecuteTemplate(w, "comment.form", commentForm)
			return
		}
		comment := issue1.Comment{
			Commenter: sess.Get(s.sessionValues.username),
			Content:   temp.Comment,
			ReplyTo:   -1,
		}
		_, err = s.Iss1C.CommentService.AddComment(uint(postID), &comment, sess.Get(s.sessionValues.restRefreshToken))
		if err != nil {
			if err == issue1.ErrAccessDenied {
				err = refreshTokenAuthOnSession(sess, s, w, r)
				if err != nil {
					return
				}
				_, err = s.Iss1C.CommentService.AddComment(uint(postID), &comment, sess.Get(s.sessionValues.restRefreshToken))
				if err != nil {
					showErrorPage(w, r)
					return
				}
			} else {
				showErrorPage(w, r)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
func postPostComments(s *Setup) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := getParametersFromRequestAsMap(r)
		postIDRaw := vars["postID"]
		postID, err := strconv.Atoi(postIDRaw)
		if err != nil || postID < 1 {
			show404Page(w, r)
			return
		}
		var p struct {
			Page    uint `json:"page"`
			PerPage uint `json:"perPage"`
			//Sorting issue1.SortCommentsBy `json:"sorting"`
		}
		err = json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, "invalid request JSON",
				http.StatusBadRequest)
			return
		}

		comments, err := s.Iss1C.CommentService.GetCommentsPaged(p.Page, p.PerPage, uint(postID))
		if err != nil {
			showErrorPage(w, r)
			return
		}
		type augmentedComment struct {
			*issue1.Comment
			Commenter *issue1.User
			Replies   []*augmentedComment
		}
		augComments := make([]*augmentedComment, 0)
		replies := make(map[int][]*augmentedComment, 0)
		for _, comment := range comments {
			augComment := augmentedComment{
				Comment: comment,
				Replies: make([]*augmentedComment, 0),
			}
			augComment.Commenter, err = s.Iss1C.UserService.GetUser(comment.Commenter)
			if err != nil {
				showErrorPage(w, r)
				return
			}
			replies[int(comment.ID)] = augComment.Replies
			augComments = append(augComments, &augComment)
		}
		boardData := make([]*augmentedComment, 0)
		for _, augComment := range augComments {
			if augComment.ReplyTo == -1 {
				boardData = append(boardData, augComment)
			} else {
				replies[augComment.ReplyTo] = append(replies[augComment.ReplyTo], augComment)
			}
		}
		_ = s.templates.ExecuteTemplate(w, "comment.board", boardData)
	}
}

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
			CSRF string
		}
		postData.CSRF, err = cSRFToken(
			"",
			s.TokenSigningSecret,
			s.CSRFTokenLifetime,
		)
		if err != nil {
			showErrorPage(w, r)
			return
		}
		err = sess.Set(s.sessionValues.csrf, postData.CSRF)
		if err != nil {
			s.Logger.Printf("server error setting value on session because: %v", err)
			showErrorPage(w, r)
			return
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
