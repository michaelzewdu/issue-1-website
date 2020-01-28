package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	gormRepo "github.com/slim-crown/issue-1-website/internal/repositories/gorm"
	"github.com/slim-crown/issue-1-website/internal/services/session"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/slim-crown/issue-1-website/internal/delivery/web"
	issue1 "github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/http.issue1"
)

func main() {

	s := web.Setup{}

	s.Logger = log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)

	var err error
	const (
		host     = "localhost"
		port     = "5432"
		dbname   = "issue#1website"
		role     = "postgres"
		password = "ntworzit"
	)
	dataSourceName := fmt.Sprintf(
		`host=%s port=%s dbname='%s' user='%s' password='%s' sslmode=disable`,
		host, port, dbname, role, password)

	db, err := gorm.Open("postgres", dataSourceName)
	if err != nil {
		s.Logger.Fatalf("database connection failed because: %s", err.Error())
	}
	defer db.Close()

	{
		if !db.HasTable(&session.Session{}) || !db.HasTable(&session.MapPair{}) {
			errs := db.AutoMigrate(&session.Session{}, &session.MapPair{}).GetErrors()
			if len(errs) > 0 {
				log.Fatalf("migration of session failed becauses: %+v", errs)
			}
		}
	}

	s.TemplatesStoragePath = "web/templates"
	s.AssetStoragePath = "web/assets"
	s.AssetServingRoute = "/assets/"

	s.HostAddress = "localhost"
	s.Port = "8081"
	s.HostAddress += ":" + s.Port

	s.CookieName = "I1Session"

	s.TokenSigningSecret = []byte("secret")
	s.CSRFTokenLifetime = 7 * time.Minute
	s.SessionIdleLifetime = 1 * time.Minute
	// s.SessionIdleLifetime = 7 * 24 * time.Hour
	s.SessionHardLifetime = 30 * 24 * time.Hour
	s.HTTPS = false

	s.Iss1C = issue1.NewClient(
		http.DefaultClient,
		&url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
		},
		s.Logger,
	)
	sessionGormRepo := gormRepo.NewSessionRepo(db)
	s.SessionService = session.NewService(&sessionGormRepo)

	//mux := web.NewMux(&s)

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			switch scanner.Text() {
			case "k":
				log.Fatalln("shutting server down...")
			case "r":
				err := s.ParseTemplates()
				if err != nil {
					log.Printf("error: template parsing failed because: %w\n warning: accessing routes now may cause fatal error.", err)
				} else {
					log.Printf("templates refreshed.")
				}
			default:
				fmt.Println("unknown command")
			}
		}
	}()

	s.Logger.Println("server running...")

	//if s.HTTPS {
	//	s.HostAddress = "https://" + s.HostAddress
	//	log.Fatal(http.ListenAndServe(":"+s.Port, mux))
	//} else {
	//	s.HostAddress = "http://" + s.HostAddress
	//	log.Fatal(http.ListenAndServe(":"+s.Port, mux))
	//}

	i1 := s.Iss1C
	stdoutLogger := s.Logger

	/*
		u, err := i1.UserService.GetUser("slimmy")
		stdoutLogger.Printf("\nGetUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)
	*/
	/*
		u, err := i1.UserService.AddUser(&issue1.User{
			Username:   "loveless",
			Email:      "stars@destination.com",
			FirstName:  "Jeff",
			MiddleName: "k.",
			LastName:   "Shoes",
			Bio:        "i don't know what's real",
			Password:   "password",
		})
		stdoutLogger.Printf("\nAddUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)
	*/
	/*
		users, err := i1.UserService.SearchUsers(issue1.PaginateParams{
			Pattern:   "",
			SortUsersBy:    issue1.SortUsersByUsername,
			SortOrder: issue1.SortDescending,
			Limit:     0,
			Offset:    0,
		})

		stdoutLogger.Printf("\nSearchUser\n - - - - value:\n%#v\n\n - - - - error:\n%+v", users, err)
		if err == nil {
			for _, u := range users {
				stdoutLogger.Printf("%v\n", u)
			}
		}
	*/
	/*
		token, err := i1.GetAuthToken("loveless","password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.Logout(token)
		stdoutLogger.Printf("\nLogout\n - - - - error:\n%+v", err)

		refreshedToken, err := i1.RefreshAuthToken(token)
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", refreshedToken, err)

	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		u, err := i1.UserService.UpdateUser(
			"loveless",
			&issue1.User{
				Bio: "i don't know what's real!",
			},
			token,
		)
		stdoutLogger.Printf("\nUpdateUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)
	*/
	/*
		u, err := i1.UserService.AddUser(&issue1.User{
			Username:   "randoWanda",
			Email:      "unod@commanda.com",
			FirstName:  "Anda",
			MiddleName: "A",
			LastName:   "Boss",
			Bio:        "i don't know what's real either",
			Password:   "password",
		})
		stdoutLogger.Printf("\nAddUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)

		token, err := i1.GetAuthToken("randoWanda", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.UserService.DeleteUser("randoWanda",token)
		stdoutLogger.Printf("\nDeleteUser\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.UserService.BookmarkPost("loveless", 3, token)
		stdoutLogger.Printf("\nBookmarkPost\n - - - - error:\n%+v", err)

		p, err := i1.UserService.GetUserBookmarks("loveless", token)
		stdoutLogger.Printf("\nGetUserBookmarks\n - - - - value:\n%#v\n\n - - - - error:\n%+v", p, err)

		err = i1.UserService.DeleteBookmark("loveless", 3, token)
		stdoutLogger.Printf("\nDeleteBookmark\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		image, err := os.Open("E:\\Files\\MuSec\\Alternative\\! My Bloody Valentine\\My Bloody Valentine [2008] Loveless\\front.jpg")
		if err != nil {
			stdoutLogger.Printf("hmm...error: %+v\n", err)
			panic(err)
		}
		defer image.Close()

		path, err := i1.UserService.AddPicture("loveless", image, "lovelessness.jpg", token)
		stdoutLogger.Printf("\nAddPicture\n - - - - value:\n%s\n\n - - - - error:\n%+v", path, err)


		//err = i1.UserService.RemovePicture("loveless", token)
		//stdoutLogger.Printf("\nRemovePicture\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		u, err := i1.FeedService.GetFeedSorting("loveless", token)
		stdoutLogger.Printf("\nGetFeedSorting\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		c, err := i1.FeedService.GetFeedSubscriptions("loveless", token, issue1.SortBySubscriptionTime, "")
		stdoutLogger.Printf("\nGetFeedSubscriptions\n - - - - value:\n%+v\n\n - - - - error:\n%+v", c, err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.FeedService.SetFeedSorting(issue1.SortNew,"loveless",token)
		stdoutLogger.Printf("\nSetFeedSorting\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.FeedService.SubscribeToChannel("loveless", "chromagnum", token)
		stdoutLogger.Printf("\nSubscribeToChannel\n - - - - error:\n%+v", err)

		err = i1.FeedService.UnsubscribeFromChannel("loveless", "chromagnum", token)
		stdoutLogger.Printf("\nUnsubscribeFromChannel\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		err = i1.FeedService.SubscribeToChannel("loveless", "chromagnum", token)
		stdoutLogger.Printf("\nSubscribeToChannel\n - - - - error:\n%+v", err)

		posts, err := i1.FeedService.GetFeedPosts("loveless", issue1.SortHot,
			issue1.PaginateParams{issue1.SortDescending,5,0,}, token)
		stdoutLogger.Printf("\nGetFeedPosts\n - - - - value:\n%+v\n\n - - - - error:\n%+v", posts, err)

		err = i1.FeedService.UnsubscribeFromChannel("loveless", "chromagnum", token)
		stdoutLogger.Printf("\nUnsubscribeFromChannel\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		r, err := i1.ReleaseService.GetReleaseAuthorized(52, token)
		stdoutLogger.Printf("\nGetRelease\n - - - - value:\n%+v\n\n - - - - error:\n%+v", *r, err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		image, err := os.Open("E:\\Files\\MuSec\\Alternative\\! My Bloody Valentine\\My Bloody Valentine [2008] Loveless\\front.jpg")
		if err != nil {
			stdoutLogger.Printf("hmm...error: %+v\n", err)
			panic(err)
		}
		defer image.Close()

		path, err := i1.ReleaseService.AddImageRelease(&issue1.Release{
			OwnerChannel: "chromagnum",
			Type:  "image",
			Metadata: issue1.Metadata{
				Title:         "Loveless Album Art",
				ReleaseDate:   time.Now(),
				GenreDefining: "Album Art",
				Description:   "Loveless album album art",
				Other: issue1.Other{
					Authors:       []string{"my bloody valentine"},
					Genres:        []string{"Shoegaze"},
				},
			},
		}, image, "lovelessness.jpg", token)
		stdoutLogger.Printf("\nAddImageRelease\n - - - - value:\n%s\n\n - - - - error:\n%+v", path, err)

		image2, err := os.Open("C:\\Users\\cosmicbridgeman\\Pictures\\The_Hunchback_of_Notre_Dame-fPKapNhofB4.ogv_snapshot_00.01.38_[2019.02.18_12.11.46].jpg")
		if err != nil {
			stdoutLogger.Printf("hmm...error: %+v\n", err)
			panic(err)
		}
		defer image2.Close()

		path, err = i1.ReleaseService.UpdateImageRelease(path.ID,&issue1.Release{
			Metadata: issue1.Metadata{
				Title:         "new stuff",
				},
		}, image2, "kordiourie.jpg", token)
		stdoutLogger.Printf("\nUpdateImageRelease\n - - - - value:\n%s\n\n - - - - error:\n%+v", path, err)


		err = i1.ReleaseService.DeleteRelease(path.ID, token)
		stdoutLogger.Printf("\nDeleteReleasev\n - - - - error:\n%+v", err)
	*/
	/*
		token, err := i1.GetAuthToken("loveless", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		r, err := i1.ReleaseService.AddTextRelease(&issue1.Release{
			OwnerChannel: "chromagnum",
			Type:         "text",
			Content:      "1 minute and 40 seconds to midnight.",
			Metadata: issue1.Metadata{
				Title:         "This is Not A Test",
				ReleaseDate:   time.Now(),
				GenreDefining: "Atomic",
				Description:   "Full stop.",
				Other: issue1.Other{
					Authors: []string{"Man"},
					Genres:  []string{"Catastrophe"},
				},
			},
		}, token)
		stdoutLogger.Printf("\nAddTextRelease\n - - - - value:\n%s\n\n - - - - error:\n%+v", r, err)

		r, err = i1.ReleaseService.UpdateRelease(r.ID, &issue1.Release{
			Content: "1 minute and 39 seconds to midnight.",
		}, issue1.Text, token)
		stdoutLogger.Printf("\nUpdateRelease\n - - - - value:\n%s\n\n - - - - error:\n%+v", r, err)

		err = i1.ReleaseService.DeleteRelease(r.ID, token)
		stdoutLogger.Printf("\nDeleteRelease\n - - - - error:\n%+v", err)
	*/
	/*
		releases, err := i1.ReleaseService.SearchReleases("", issue1.SortByType,
			issue1.PaginateParams{
			SortOrder: issue1.SortDescending,
			Limit:     2,
			Offset:    0,
		})

		stdoutLogger.Printf("\nSearchReleases\n - - - - value:\n%#v\n\n - - - - error:\n%+v", releases, err)
		if err == nil {
			for _, u := range releases {
				stdoutLogger.Printf("%v\n", u)
			}
		}
	*/
	/*
		token, err := i1.GetAuthToken("slimmy", "password")
		stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

		comment, err := i1.CommentService.AddComment(5, &issue1.Comment{
			Content: "Where did the good things go?",
		}, token)
		stdoutLogger.Printf("\nAddComment\n - - - - value:\n%s\n\n - - - - error:\n%+v", comment, err)

		comment, err = i1.CommentService.GetComment(comment.ID, comment.OriginPost)
		stdoutLogger.Printf("\nGetComment\n - - - - value:\n%+v\n\n - - - - error:\n%+v", comment, err)

		comment, err = i1.CommentService.UpdateComment(comment.ID, comment.OriginPost, &issue1.Comment{
			Content: "I want peace.",
		}, token)
		stdoutLogger.Printf("\nUpdateComment\n - - - - value:\n%s\n\n - - - - error:\n%+v", comment, err)

		reply, err := i1.CommentService.AddReply(comment.ID, comment.OriginPost, &issue1.Comment{
			Content: "Give me a week",
		}, token)
		stdoutLogger.Printf("\nAddReply\n - - - - value:\n%s\n\n - - - - error:\n%+v", reply, err)

		comments, err := i1.CommentService.GetCommentsPaged(1, 5, comment.OriginPost)
		stdoutLogger.Printf("\nGetCommentsPaged\n - - - - value:\n%#v\n\n - - - - error:\n%+v", comments, err)
		if err == nil {
			for _, u := range comments {
				stdoutLogger.Printf("%v\n", u)
			}
		}

		comments, err = i1.CommentService.GetRepliesPaged(1, 5, comment.ID, comment.OriginPost)
		stdoutLogger.Printf("\nGetRepliesPaged\n - - - - value:\n%#v\n\n - - - - error:\n%+v", comments, err)
		if err == nil {
			for _, u := range comments {
				stdoutLogger.Printf("%v\n", u)
			}
		}

		err = i1.CommentService.DeleteComment(comment.ID, comment.OriginPost, token)
		stdoutLogger.Printf("\nDeleteRelease\n - - - - error:\n%+v", err)

		err = i1.CommentService.DeleteComment(reply.ID, reply.OriginPost, token)
		stdoutLogger.Printf("\nDeleteRelease\n - - - - error:\n%+v", err)
	*/
	//channels,err:= i1.ChannelService.GetChannelPosts("Isis Cane")
	//if err == nil {
	//	for _, u := range channels {
	//		stdoutLogger.Printf("%v\n", u)
	//	}
	//}
	//token, err := i1.GetAuthToken("Isis kane", "password")
	//stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)
	//c, err := i1.ChannelService.AddChannel(&issue1.Channel{
	//			ChannelUsername :"cbezdojsdnhugphos",
	//			Name:"name",
	//			Description:"description",
	//	 },token)
	//
	//	 stdoutLogger.Printf("\nAddChannel\n - - - - value:\n%+v\n\n - - - - error:\n%+v", c, err)

	//ca, err := i1.ChannelService.GetOwner("Isis Kane", token)
	//if err == nil {
	//	for _, u := range ca {
	//		stdoutLogger.Printf("%v\n", u)
	//	}
	//}
	//c, err = i1.ChannelService.GetAdmins("bob", token)
	//if err == nil {
	//	for _, u := range c {
	//		stdoutLogger.Printf("%v\n", u)
	//	}
	//}

	//ca, err := i1.ChannelService.GetOwner("Isis kane",token)
	//stdoutLogger.Printf(" %v\n",ca)
	//cB, err := i1.ChannelService.GetChannelAuthorized("Isis kane",token)
	//stdoutLogger.Printf("\nGetChannel\n - - - - value:\n%+v\n\n - - - - error:\n%+v", cB)
	////cb, err := i1.ChannelService.GetChannelPosts("IsisCane",96)
	////stdoutLogger.Printf("\nGetChannel\n - - - - value:\n%+v\n\n - - - - error:\n%+v", cb)
	//cc, err := i1.ChannelService.GetChannelPosts("Isis kane")
	//if err == nil {
	//	for _, u := range cc {
	//		stdoutLogger.Printf("%v\n", u)
	//	}
	//}
	/*
		p, err := i1.PostService.AddPost(&issue1.Post{
			PostedByUsername: "slimmy",
			OriginChannel:    "chromagnum",
			Title:            "Lorem ipsum dolor sit amet",
			Description:      "consectetur adipisicing elit, sed do eiusmodtempor",
		},tokenUser)
		stdoutLogger.Printf("\nAddPost\n - - - - value:\n%+v\n\n - - - - error:\n%+v", p, err)
	*/
	/*
		p, err := i1.PostService.GetPost(9)
		stdoutLogger.Printf("\nGetPost\n - - - - value:\n%+v\n\n - - - - error:\n%+v", p, err)
	*/

	/*

		p, err := i1.PostService.UpdatePost(7, &issue1.Post{
			Description: "Amharic,English,French,Spanish",
		}, token)
		stdoutLogger.Printf("\nUpdatePost\n - - - - value:\n%+v\n\n - - - - error:\n%+v", p, err)
	*/

	/*
		err = i1.PostService.DeletePost(7,token)
		stdoutLogger.Printf("\nDeletePost\n - - - - error:\n%+v", err)
	*/

	/* NOT USED, USING COMMENT SERVICE INSTEAD
	comments, err:= i1.PostService.GetPostComments(5)
	stdoutLogger.Printf("\nGetPostComments\n - - - - value:\n%+v\n\n - - - - error:\n%+v", comments, err)
	if err== nil{
		for _,u:= range comments{
			stdoutLogger.Printf("%v\n", u)
		}
	}
	*/

	releases, err := i1.PostService.GetPostReleases(4)
	stdoutLogger.Printf("\nGetPostReleases\n - - - - value:\n%+v\n\n - - - - error:\n%+v", releases, err)
	if err == nil {
		for _, u := range releases {
			stdoutLogger.Printf("%v\n", u)
		}
	}

	/*
		star, err := i1.PostService.GetPostStar(9,"loveless")
		stdoutLogger.Printf("\nGetPostStar\n - - - - value:\n%+v\n\n - - - - error:\n%+v", star, err)
	*/
	/*
		star, err := i1.PostService.UpdatePostStar(7, &issue1.Star{
			Username:   "loveless",
			NumOfStars: 3,
		}, token)
		stdoutLogger.Printf("\nUpdatePostStar\n - - - - value:\n%+v\n\n - - - - error:\n%+v", star, err)
	*/

	/*stars, err:= i1.PostService.GetPostStars(7)
	stdoutLogger.Printf("\nGetPostStars\n - - - - value:\n%+v\n\n - - - - error:\n%+v", stars, err)
	if err== nil{
		for _,u:= range stars{
			stdoutLogger.Printf("%v\n", u)
		}
	}
	*/

	posts, err := i1.PostService.GetPosts(1, 7)
	if err == nil {
		for _, u := range posts {
			stdoutLogger.Printf("%v\n", u)
		}
	}
}
