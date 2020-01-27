package issue1

import (
	"encoding/json"
	"fmt"

	//"io"
	"net/http"
	"net/url"
	"strconv"
)

// PostService is used to interact with the user services on the REST server.
type PostService service

// SortPostsBy  holds enums used by to specify the attribute entities are sorted with
type SortPostsBy string

// PostSorting constants used by SearchPosts methods
const (
	SortPostsByCreationTime SortPostsBy = "creation_time"
	SortPostsByChannel      SortPostsBy = "channel_from"
	SortPostsByPoster       SortPostsBy = "posted_by"
	SortPostsByTitle        SortPostsBy = "title"
)

// GetPosts returns a list of all posts. They are sorted according to the
// default sorting on the REST server. To specify sorting, user Searchposts and
// user an empty string for the pattern.
func (s *PostService) GetPosts(page, perPage uint) ([]*Post, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchPosts("", "", p)
}

// SearchPostsPaged is a utility wrapper for SearchPosts for easy pagination,
func (s *PostService) SearchPostsPaged(page, perPage uint, pattern string, by SortPostsBy, order SortOrder) ([]*Post, error) {
	p := PaginateParams{
		SortOrder: order,
	}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchPosts(pattern, by, p)
}

// SearchPosts returns a list of Post according to the passed in parameters.
// An empty pattern matches all posts. If any of the fields on the passed in
// PaginateParams are omitted, it'll use the default values.
func (s *PostService) SearchPosts(pattern string, by SortPostsBy, params PaginateParams) ([]*Post, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts")
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
		default:
			return nil, ErrRESTServerError
		}
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
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

// GetPost returns the post under the given id.
func (s *PostService) GetPost(id uint) (*Post, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d", id)
	)
	req := s.client.newRequest(path, method)
	js, _, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		return nil, ErrPostNotFound
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}
	p := new(Post)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &p)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return p, nil
}

//AddPost creates and returns post and an error if found any
func (s *PostService) AddPost(p *Post, authToken string) (*Post, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("/posts")
	)
	req := s.client.newRequest(path, method)
	err := addBodyToRequestAsJSON(req, p)

	if err != nil {
		return nil, err
	}
	addJWTToRequest(req, authToken)

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
			return nil, ErrInvalidData
		default:

		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}
	p = new(Post)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &p)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return p, nil

}

// DeletePost removes the post under the given post id.
func (s *PostService) DeletePost(id uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/posts/%d", id)
	)
	req := s.client.newRequest(path, method)

	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return err
	}

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
			return ErrPostNotFound
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusBadRequest:
			return ErrInvalidData
		default:
			return ErrRESTServerError
		}
	case "error":
		fallthrough
	default:
		return ErrRESTServerError
	}
	return nil
}

// UpdatePost removes the post under the given post id.
func (s *PostService) UpdatePost(id uint, p *Post, authToken string) (*Post, error) {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/posts/%d", id)
	)
	req := s.client.newRequest(path, method)
	err := addBodyToRequestAsJSON(req, p)

	if err != nil {
		return nil, err
	}
	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}
	s.client.Logger.Printf("\n%+v\n", js)
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
			return nil, ErrPostNotFound
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		default:
			return nil, ErrRESTServerError
		}
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}
	post := new(Post)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &post)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return post, nil
}

// GetPostComments returns the comments under the given id post.
func (s *PostService) GetPostComments(id uint) ([]*Comment, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d/comments", id)
	)
	req := s.client.newRequest(path, method)
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			return nil, ErrPostNotFound
		default:
		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}
	// fmt.Println("\n1.hfehelh")
	comments := make([]*Comment, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &comments)
	// fmt.Printf("\n2.value:\n%+v", data)
	if err != nil {
		fmt.Printf("%s", err.Error())
		// fmt.Printf("\n3.value:\n%+v\n", comments)
		return nil, ErrRESTServerError
	}

	return comments, nil
}

// GetPostReleases returns the releases under the given id post.
func (s *PostService) GetPostReleases(id uint) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d/releases", id)
	)
	req := s.client.newRequest(path, method)
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			return nil, ErrPostNotFound
		default:
		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}

	releases := make([]*Release, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &releases)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return releases, nil
}

// GetPostStars returns the stars of the given id post.
func (s *PostService) GetPostStars(id uint) ([]*Star, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d/stars", id)
	)
	req := s.client.newRequest(path, method)
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			return nil, ErrPostNotFound
		default:
		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}

	stars := make([]*Star, 0)
	// fmt.Printf("hehlel")
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &stars)
	if err != nil {
		// fmt.Printf("%s", err.Error())
		return nil, ErrRESTServerError
	}
	return stars, nil
}

// GetPostStar returns the stars of the given id post.
func (s *PostService) GetPostStar(id uint, username string) (*Star, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d/stars/%s", id, username)
	)
	req := s.client.newRequest(path, method)
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			return nil, ErrStarNotFound
		default:
		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}

	star := new(Star)
	// fmt.Printf("hehlel")
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &star)
	if err != nil {
		// fmt.Printf("%s", err.Error())
		return nil, ErrRESTServerError
	}
	return star, nil
}

// UpdatePostStar returns the stars of the updated star post.
func (s *PostService) UpdatePostStar(id uint, st *Star, authToken string) (*Star, error) {
	var (
		method = http.MethodPut
		path   = fmt.Sprintf("/posts/%d/stars", id)
	)
	req := s.client.newRequest(path, method)
	err := addBodyToRequestAsJSON(req, st)

	if err != nil {
		return nil, err
	}
	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		switch statusCode {
		case http.StatusBadRequest:
			return nil, ErrInvalidData
		case http.StatusNotFound:
			fmt.Printf("%s", err.Error())
			return nil, ErrStarNotFound
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		default:
		}
		fallthrough
	case "error":
		fallthrough
	default:
		return nil, ErrRESTServerError
	}

	star := new(Star)
	// fmt.Printf("hehlel")
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &star)
	if err != nil {
		// fmt.Printf("%s", err.Error())
		return nil, ErrRESTServerError
	}
	return star, nil
}
