package issue1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// CommentService is used to interact with the comment service on the REST server.
type CommentService service

// SortCommentsBy holds enums that specify by what attribute comments are sorted
type SortCommentsBy string

// Sorting constants used by SearchComment methods
const (
	SortCommentsByCreationTime SortCommentsBy = "creation_time"
)

func (s *CommentService) AddComment(postID uint, c *Comment, authToken string) (*Comment, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("posts/%d/comments", postID)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	c.OriginPost = postID
	err := addBodyToRequestAsJSON(req, c)
	if err != nil {
		return nil, err
	}
	return s.addComment(req)
}

func (s *CommentService) AddReply(commentID, postID uint, c *Comment, authToken string) (*Comment, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("posts/%d/comments/%d/replies", postID, commentID)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	//c.OriginPost = postID
	c.ReplyTo = int(commentID)
	err := addBodyToRequestAsJSON(req, c)
	if err != nil {
		return nil, err
	}
	return s.addComment(req)
}

func (s *CommentService) addComment(req *http.Request) (*Comment, error) {
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
		case http.StatusNotFound:
			switch jF.ErrorReason {
			case "postID":
				return nil, ErrPostNotFound
			case "commentID":
				return nil, ErrCommentNotFound
			case "username":
				return nil, ErrUserNotFound
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

	newComment := new(Comment)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, newComment)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return newComment, nil
}

func (s *CommentService) GetComment(id, postId uint) (*Comment, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/posts/%d/comments/%d", postId, id)
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
		return nil, ErrCommentNotFound
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
	c := new(Comment)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, c)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return c, nil
}

func (s *CommentService) GetCommentsPaged(page, perPage, postID uint) ([]*Comment, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.GetComments(postID, "", p)

}

func (s *CommentService) GetRepliesPaged(page, perPage, commentID, postID uint) ([]*Comment, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.GetReplies(commentID, postID, "", p)
}

func (s *CommentService) GetComments(postID uint, by SortCommentsBy, params PaginateParams) ([]*Comment, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("posts/%d/comments", postID)
	)

	queries := url.Values{}
	// use params only if not default values
	if params.Limit != 0 || params.Offset != 0 {
		queries.Set("limit", strconv.FormatUint(uint64(params.Limit), 10))
		queries.Set("offset", strconv.FormatUint(uint64(params.Offset), 10))
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
	return s.getComments(req)
}

func (s *CommentService) GetReplies(commentID, postID uint, by SortCommentsBy, params PaginateParams) ([]*Comment, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("posts/%d/comments/%d/replies", postID, commentID)
	)

	queries := url.Values{}
	// use params only if not default values
	if params.Limit != 0 || params.Offset != 0 {
		queries.Set("limit", strconv.FormatUint(uint64(params.Limit), 10))
		queries.Set("offset", strconv.FormatUint(uint64(params.Offset), 10))
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
	return s.getComments(req)
}

func (s *CommentService) getComments(req *http.Request) ([]*Comment, error) {
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

	comments := make([]*Comment, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &comments)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return comments, nil
}

func (s *CommentService) UpdateComment(id, postId uint, c *Comment, authToken string) (*Comment, error) {
	if c.Content == "" {
		return nil, ErrInvalidData
	}
	var (
		method = http.MethodPatch
		path   = fmt.Sprintf("/posts/%d/comments/%d", postId, id)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, c)
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
			return nil, ErrCommentNotFound
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

	newComment := new(Comment)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, newComment)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return newComment, nil
}

func (s *CommentService) DeleteComment(id, postId uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/posts/%d/comments/%d", postId, id)
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
		return ErrRESTServerError
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
