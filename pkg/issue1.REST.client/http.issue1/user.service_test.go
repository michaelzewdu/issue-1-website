package issue1

import (
	"testing"
	"time"
)

var aUser = User{"slimmy", "miku@miku.com", "mikee", "boy", "shifi", time.Now(),
	"like you don't know", "addis", "abc123"}

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
