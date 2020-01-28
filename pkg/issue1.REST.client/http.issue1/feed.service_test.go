package issue1

import "testing"

func TestGetFeedSorting(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	u, err := i1.FeedService.GetFeedSorting("loveless", token)
	stdoutLogger.Printf("\nGetFeedSorting\n - - - - value:\n%+v\n\n - - - - error:\n%+v", u, err)

	if err != nil {
		t.Fatal(err)
	}
}
func TestGetFeedSubscription(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	c, err := i1.FeedService.GetFeedSubscriptions("loveless", token, SortBySubscriptionTime, "")
	stdoutLogger.Printf("\nGetFeedSubscriptions\n - - - - value:\n%+v\n\n - - - - error:\n%+v", c, err)

	if err != nil {
		t.Fatal(err)
	}
}
func TestSetFeedSorting(t *testing.T) {
	token, err := i1.GetAuthToken("loveless", "password")
	stdoutLogger.Printf("\nGetAuthToken\n - - - - value:\n%#v\n\n - - - - error:\n%+v", token, err)

	err = i1.FeedService.SetFeedSorting(SortNew, "loveless", token)
	stdoutLogger.Printf("\nSetFeedSorting\n - - - - error:\n%+v", err)

	if err != nil {
		t.Fatal(err)
	}
}
