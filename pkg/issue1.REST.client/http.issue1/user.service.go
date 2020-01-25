package issue1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// UserService is used to interact with the user services on the REST server.
type UserService service

// SortUsersBy  holds enums used by to specify the attribute entities are sorted with
type SortUsersBy string

// FeedSorting constants used by SearchUsers methods
const (
	SortUsersByCreationTime SortUsersBy = "creation_time"
	SortUsersByUsername     SortUsersBy = "username"
	SortUsersByFirstName    SortUsersBy = "first-name"
	SortUsersByLastName     SortUsersBy = "last-name"
)

// ErrUserNameOccupied is returned when the the username specified is occupied
var ErrUserNameOccupied = fmt.Errorf("user name is occupied")

// ErrEmailIsOccupied is returned when the the email specified is occupied
var ErrEmailIsOccupied = fmt.Errorf("email is occupied")

// AddUser sends a a request to create a user based on the passed in struct to the
// REST server. Returns ErrInvalidData if the struct has unacceptable data.
func (s *UserService) AddUser(u *User) (*User, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("/users")
	)
	req := s.client.newRequest(path, method)
	err := addBodyToRequestAsJSON(req, u)
	if err != nil {
		return nil, err
	}
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
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
			return nil, ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusConflict:
			switch jF.ErrorReason {
			case "email":
				return nil, ErrEmailIsOccupied
			case "username":
				return nil, ErrUserNameOccupied
			default:
			}
			fallthrough
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

	newUser := new(User)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, newUser)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return newUser, nil
}

// GetUser returns the user under the given username. To get private info of a
// user, user GetUserAuthorized.
func (s *UserService) GetUser(username string) (*User, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s", username)
	)
	req := s.client.newRequest(path, method)
	return s.getUser(req)
}

// GetUserAuthorized gets the user under the given username including their email
// and other private information.
func (s *UserService) GetUserAuthorized(username, token string) (*User, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s", username)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, token)
	return s.getUser(req)
}

func (s *UserService) getUser(req *http.Request) (*User, error) {
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		return nil, ErrUserNotFound
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
	u := new(User)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, u)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return u, nil
}

// GetUsers returns a list of all users. They are sorted according to the
// default sorting on the REST server. To specify sorting, user SearchUsers and
// user an empty string for the pattern.
func (s *UserService) GetUsers(page, perPage uint) ([]*User, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchUsers("", "", p)
}

// SearchUsersPaged is a utility wrapper for SearchUsers for easy pagination,
func (s *UserService) SearchUsersPaged(page, perPage uint, pattern string, by SortUsersBy, order SortOrder) ([]*User, error) {
	p := PaginateParams{
		SortOrder: order,
	}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchUsers(pattern, by, p)
}

// SearchUsers returns a list of user according to the passed in parameters.
// An empty pattern matches all users. If any of the fields on the passed in
// PaginateParams are omitted, it'll use the default values.
func (s *UserService) SearchUsers(pattern string, by SortUsersBy, params PaginateParams) ([]*User, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users")
	)

	queries := url.Values{}
	// use params only if not default values
	if params.Limit != 0 || params.Offset != 0 {
		queries.Set("limit", strconv.FormatUint(uint64(params.Limit), 10))
		queries.Set("offset", strconv.FormatUint(uint64(params.Offset), 10))
	}
	if pattern != "" {
		queries.Set("pattern", url.QueryEscape(pattern))
	}
	if by != "" {
		var qString string
		if params.SortOrder != "" {
			qString = fmt.Sprintf("%s_%s", by, params.SortOrder)
		} else {
			qString = string(by)
		}
		queries.Set("sort", qString)
	}

	req := s.client.newRequest(path, method)
	req.URL.RawQuery = queries.Encode()

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
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "limit":
				fallthrough
			case "offset":
				fallthrough
			default:
			}
			fallthrough
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

	users := make([]*User, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &users)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return users, nil
}

// UpdateUser updates the user under the given username based on the passed in struct.
// When changing username, be sure to get new tokens after this call as the one used
// here won't work.
func (s *UserService) UpdateUser(username string, u *User, authToken string) (*User, error) {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/users/%s", username)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, u)
	if err != nil {
		return nil, err
	}
	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
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
			return nil, ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			return nil, ErrUserNotFound
		case http.StatusConflict:
			switch jF.ErrorReason {
			case "email":
				return nil, ErrEmailIsOccupied
			case "username":
				return nil, ErrUserNameOccupied
			default:
			}
			fallthrough
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

	newUser := new(User)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, newUser)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return newUser, nil
}

// DeleteUser removes the user under the given username.
func (s *UserService) DeleteUser(username, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/users/%s", username)
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
	case "fail":
		jF, ok := js.Data.(*jSendFailData)
		if !ok {
			return ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			return ErrInvalidData
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// GetUserBookmarks gets the posts that have been bookmarked by the user of the given auth token.
func (s *UserService) GetUserBookmarks(username string, authToken string) (map[time.Time]*Post, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/users/%s/bookmarks", username)
	)
	req := s.client.newRequest(path, method)

	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
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
			return nil, ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusConflict:
			switch jF.ErrorReason {
			case "email":
				return nil, ErrEmailIsOccupied
			case "username":
				return nil, ErrUserNameOccupied
			default:
			}
			fallthrough
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

	bookmarks := new(map[time.Time]*Post)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, bookmarks)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return *bookmarks, nil
}

// BookmarkPost adds the post under the given ID to the bookmark list of the user
// under the given username.
func (s *UserService) BookmarkPost(username string, postID int, authToken string) error {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/users/%s/bookmarks/%d", username, postID)
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
			case "postID":
				return ErrPostNotFound
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

// DeleteBookmark removes the given postID from the user's bookmark list.
func (s *UserService) DeleteBookmark(username string, postID int, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/users/%s/bookmarks/%d", username, postID)
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

// AddPicture sets the passed in image as the user's picture for the user
// under the passed in username.
func (s *UserService) AddPicture(username string, image io.Reader, imageName, authToken string) (string, error) {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/users/%s/picture", username)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	err := addImageToRequest(req, image, imageName)
	if err != nil {
		return "", err
	}
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return "", err
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
			return "", ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "image":
				return "", ErrUnacceptedImageType
			case "username":
				return "", ErrPostNotFound
			}
		case http.StatusNotFound:
			return "", ErrUserNotFound
		default:
			return "", ErrRESTServerError
		}
	case "error":
		return "", ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return "", ErrAccessDenied
		default:
			return "", ErrRESTServerError
		}
	}
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return "", ErrRESTServerError
	}

	var imageURL string
	err = json.Unmarshal(*data, &imageURL)
	if err != nil {
		return "", ErrRESTServerError
	}
	return imageURL, nil
}

// RemovePicture picture removes the picture of the user under the given
// username.
func (s *UserService) RemovePicture(username, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/users/%s/picture", username)
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
