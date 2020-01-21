package issue1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FeedService is used to interact with the Feed service on the REST server.
type FeedService service

// FeedSorting enums used specify how posts in a Feed are sorted
type FeedSorting string

// SortSubscriptionsBy enums used specify how subscribed to channels are sorted
// when retrieved.
type SortSubscriptionsBy string

// FeedSorting constants used by SearchUsers methods
const (
	// SortTop sorts the posts according to their star count
	SortTop FeedSorting = "top"
	// SortHot sorts the posts according to their comment count
	SortHot FeedSorting = "hot"
	// SortNew sorts the posts according to their creation time
	SortNew FeedSorting = "new"
	// NotSet signifies sort hasn't been set
	NotSet FeedSorting = ""

	// SortByUsername orders channels according to their username
	SortByChannelsByUsername SortSubscriptionsBy = "username"
	// SortByName orders channels according to their name
	SortChannelsByName SortSubscriptionsBy = "name"
	// SortBySubscriptionTime orders channels according to the time the Feed subscribed to them
	SortBySubscriptionTime SortSubscriptionsBy = "sub-time"
)

// GetFeedSorting returns the sorting setting for the feed of the given user.
func (s *FeedService) GetFeedSorting(username, token string) (FeedSorting, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s/feed", username)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, token)
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return NotSet, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		return NotSet, ErrUserNotFound
	case "error":
		return NotSet, ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return NotSet, ErrAccessDenied
		case http.StatusInternalServerError:
			fallthrough
		default:
			return NotSet, ErrRESTServerError
		}
	}
	f := new(Feed)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return NotSet, ErrRESTServerError
	}
	err = json.Unmarshal(*data, f)
	if err != nil {
		return NotSet, ErrRESTServerError
	}
	return f.Sorting, nil
}

// GetFeedPostsPaged is a utility wrapper for GetFeedPosts for easy pagination.
func (s *FeedService) GetFeedPostsPaged(page, perPage uint, username, token string, sorting FeedSorting) ([]*Post, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.GetFeedPosts(username, token, sorting, p)
}

// GetFeedPosts returns a list of posts from the given's user feed sorted according
// to the the passed FeedSorting. If any of the fields on the passed in PaginateParams are
// omitted, it'll use the default values.
func (s *FeedService) GetFeedPosts(username, token string, sorting FeedSorting, params PaginateParams) ([]*Post, error) {

	// TODO test

	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s/feed/posts", username)
	)

	queries := url.Values{}
	// use params only if not default values
	if params.Limit != 0 || params.Offset != 0 {
		queries.Set("limit", strconv.FormatUint(uint64(params.Limit), 10))
		queries.Set("offset", strconv.FormatUint(uint64(params.Offset), 10))
	}
	if sorting != "" {
		queries.Set("sort", fmt.Sprintf("%s", sorting))
	}

	req := s.client.newRequest(path, method)
	req.URL.RawQuery = queries.Encode()
	addJWTToRequest(req, token)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		jF, ok := js.Data.(*jSendFailData)
		if !ok {
			return nil, ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusNotFound:
			return nil, ErrUserNotFound
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "limit":
				fallthrough
			case "offset":
				fallthrough
			default:
			}
		default:
			return nil, ErrRESTServerError
		}
	case "error":
		return nil, ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}

	posts := make([]*Post, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &posts)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return posts, nil
}

// GetFeedSubscriptions returns a list of channels from the given's user feed sorted according
// to the the passed SortSubscriptionsBy. If any of the fields on the passed in PaginateParams are
// omitted, it'll use the default values.
func (s *FeedService) GetFeedSubscriptions(username, token string, by SortSubscriptionsBy, order SortOrder) (map[time.Time]*Channel, error) {

	// TODO test

	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s/feed/channels", username)
	)

	queries := url.Values{}
	// use params only if not default values
	if by != "" {
		var qString string
		if order != "" {
			qString = fmt.Sprintf("%s_%s", by, order)
		} else {
			qString = string(by)
		}
		queries.Set("sort", qString)
	}

	req := s.client.newRequest(path, method)
	req.URL.RawQuery = queries.Encode()
	addJWTToRequest(req, token)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		jF, ok := js.Data.(*jSendFailData)
		if !ok {
			return nil, ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusNotFound:
			return nil, ErrUserNotFound
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "limit":
				fallthrough
			case "offset":
				fallthrough
			default:
			}
		default:
			return nil, ErrRESTServerError
		}
	case "error":
		return nil, ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}

	channels := make([]*struct {
		Channelname      string    `json:"channelname"`
		Name             string    `json:"name"`
		SubscriptionTime time.Time `json:"subscriptionTime"`
	}, 0)

	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &channels)
	if err != nil {
		return nil, ErrRESTServerError
	}
	out := make(map[time.Time]*Channel, 0)
	for _, channel := range channels {
		// TODO return actual channels
		out[channel.SubscriptionTime] = &Channel{
			Username: channel.Channelname,
			Name:     channel.Name,
		}
	}
	return out, nil
}

// SubscribeToChannel adds the channel under the given name to the  list of the channels
// that aggregates into a the given user's feed.
func (s *FeedService) SubscribeToChannel(username, channelname string, authToken string) error {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("/users/%s/feed/channels", username)
	)
	req := s.client.newRequest(path, method)

	addJWTToRequest(req, authToken)

	err := addBodyToRequestAsJSON(req, struct {
		Channelname string `json:"channelname"`
	}{channelname})
	if err != nil {
		return err
	}

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return err
	}

	// the following ugly piece of code will return ServerError
	// in most most cases. Most will never ever happen but that's
	// programming for you.
	switch js.Status {
	case "success":
		break
	case "fail":
		jF, ok := js.Data.(*jSendFailData)
		if !ok {
			return ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusNotFound:
			switch jF.ErrorReason {
			case "username":
				return ErrUserNotFound
			case "channelname":
				return ErrChannelNotFound
			}
			return ErrUserNotFound
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// SetFeedSorting sets the default sorting method for the feed of the given user.
func (s *FeedService) SetFeedSorting(sorting FeedSorting, username, authToken string) error {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/users/%s/feed", username)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	err := addBodyToRequestAsJSON(req, struct {
		Sorting string `json:"defaultSorting"`
	}{string(sorting)})
	if err != nil {
		return err
	}

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return err
	}

	// the following ugly piece of code will return ServerError
	// in most most cases. Most will never ever happen but that's
	// programming for you.
	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusNotFound:
			return ErrUserNotFound
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// UnsubscribeFromChannel removes the channel from the feed of the user under the given
// username.
func (s *FeedService) UnsubscribeFromChannel(username, channelname string, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/users/%s/feed/channels/%s", username, channelname)
	)
	req := s.client.newRequest(path, method)

	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return err
	}

	// the following ugly piece of code will return ServerError
	// in most most cases. Most will never ever happen but that's
	// programming for you.
	switch js.Status {
	case "success":
		break
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		default:
			return ErrRESTServerError
		}
	}
	return nil
}
