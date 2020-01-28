package issue1

import (
	"github.com/slim-crown/issue-1-website/internal/delivery/web"
	"os"
	"testing"
	"time"
)

var aUser = User{"slimmy", "miku@miku.com", "mikee", "boy", "shifi",
	time.Now(), "like you don't know", "addis", "abc123"}

var s = web.Setup{}
var i1 = s.Iss1C
var stdoutLogger = s.Logger
var defaultUsername = "slimmy"

func TestAddUser(t *testing.T) {
	us1 := UserService{}
	wantResult := "slimmy"
	gotResult, err := us1.AddUser(&aUser)

	if err != nil {
		t.Errorf("AddUser() returned error %s", err)
	}

	if wantResult != gotResult.Username {
		t.Errorf("Wrong result recevied")
	}
}

func TestGetUser(t *testing.T) {
	us1 := UserService{}
	value := "slimmy"
	gotResult, err := us1.GetUser(value)

	if err != nil {
		t.Errorf("GetUser() returned error %s", err)
	}

	if value != gotResult.Username {
		t.Errorf("wanted %  got %", value, gotResult)
	}
}

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

func TestSearchUsers(t *testing.T) {
	var users []*User
	users, err := i1.UserService.SearchUsers(defaultUsername, SortUsersByUsername, PaginateParams{
		SortDescending,
		0,
		0,
	})

	stdoutLogger.Printf("\nSearchUser\n - - - - value:\n%#v\n\n - - - - error:\n%+v", users, err)
	if err == nil {
		for _, u := range users {
			stdoutLogger.Printf("%v\n", u)

		}
	}
	if err != nil {
		t.Fatal(err)
	}
	/*var s []string
	var s2 *User
	for _, v:=range users{
		//s=append(s,v)
		s2+=v
	}
	if s[0] !=  aUser.Username{}
	*/

}

/*
	token, err := i1.GetAuthToken("loveless","password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.Logout(token)
	stdoutLogger.Printf("\nLogout\n - - - - error:\n%+v", err)

	refreshedToken, err := i1.RefreshAuthToken(token)
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", refreshedToken, err)

*/
func TestUpdateUser(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	gotValue, err := i1.UserService.UpdateUser("loveless", &aUser, token)
	stdoutLogger.Printf("\nUpdateUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", gotValue, err)
	if gotValue.Username != "loveless" {
		t.Errorf("wanted value is 'loveless' got value is %s", gotValue.Username)
	}
	if err != nil {
		t.Fatal(err)
	}
}
func TestDeleteUser(t *testing.T) {
	u, err := i1.UserService.AddUser(&aUser)
	stdoutLogger.Printf("\nAddUser\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)

	token, err := i1.GetAuthToken("slimmy", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.UserService.DeleteUser("slimmy", token)
	stdoutLogger.Printf("\nDeleteUser\n - - - - error:\n%+v", err)

	if err != nil {
		t.Fatal(err)
	}
}

func TestBookmarkPost(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.UserService.BookmarkPost("loveless", 3, token)
	stdoutLogger.Printf("\nBookmarkPost\n - - - - error:\n%+v", err)

	if err != nil {
		t.Fatal(err)
	}

}
func TestDeleteBookmark(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.UserService.DeleteBookmark("loveless", 3, token)
	stdoutLogger.Printf("\nDeleteBookmark\n - - - - error:\n%+v", err)

	if err != nil {
		t.Fatal(err)
	}
}

func TestAddPicture(t *testing.T) {
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
}

func TestRemovePicture(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	err = i1.UserService.RemovePicture("loveless", token)
	stdoutLogger.Printf("\nRemovePicture\n - - - - error:\n%+v", err)
	if err != nil {
		t.Fatal(err)
	}
}
