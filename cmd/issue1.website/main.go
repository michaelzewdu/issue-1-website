package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/slim-crown/issue-1-website/internal/delivery/web"
	"github.com/slim-crown/issue-1-website/pkg/issue1.REST.client/issue1"
)

func main() {

	s := web.Setup{}

	s.TemplatesStoragePath = "web/templates/*"
	s.AssetStoragePath = "web/assets"
	s.AssetServingRoute = "/assets/"

	s.HostAddress = "http://localhost"
	s.Port = "8081"
	s.HostAddress += ":" + s.Port

	s.TokenAccessLifetime = 0
	s.TokenRefreshLifetime = 0
	s.TokenSigningSecret = nil

	s.Logger = log.New(os.Stdout, "", log.Lmicroseconds|log.Lshortfile)

	s.Iss1C = issue1.NewClient(
		http.DefaultClient,
		&url.URL{
			Scheme: "http",
			Host:   "localhost:8080",
		},
		s.Logger,
	)

	mux := web.NewMux(&s)

	log.Println("server starting..")

	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			switch scanner.Text() {
			case "kill":
				log.Fatalln("shutting server down...")
			default:
				fmt.Println("unkown command")
			}
		}
	}()

	log.Fatal(http.ListenAndServe(":"+s.Port, mux))

	i1 := s.Iss1C
	stdoutLogger := s.Logger

	/*
		u, err := i1.UserService.GetUser("slimmyNewNewNew")
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

	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.FeedService.SubscribeToChannel("loveless", "chromagnum", token)
	stdoutLogger.Printf("\nSubscribeToChannel\n - - - - error:\n%+v", err)

	err = i1.FeedService.UnsubscribeFromChannel("loveless", "chromagnum", token)
	stdoutLogger.Printf("\nUnsubscribeFromChannel\n - - - - error:\n%+v", err)

}
