package issue1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// SearchService is used to interact with the search service on the REST server.
type SearchService service

// SortBy holds enums that specify by what attribute comments are sorted
type SortResultsBy string

// Sorting constants used by SearchUser methods
const (
	SortByCreationTime SortResultsBy = "creation_time"
	SortByRank         SortResultsBy = "rank"
)

type SearchResults struct {
	Posts    []*Post    `json:"Posts"`
	Releases []*Release `json:"Releases"`
	Comments []*Comment `json:"Comments"`
	Channels []*Channel `json:"Channels"`
	Users    []*User    `json:"Users"`
}

func (s *SearchService) Search(pattern string, by SortResultsBy, params PaginateParams) (*SearchResults, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/search")
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
	var resp struct {
		Posts    *json.RawMessage `json:"Posts"`
		Releases *json.RawMessage `json:"Releases"`
		Comments *json.RawMessage `json:"Comments"`
		Channels *json.RawMessage `json:"Channels"`
		Users    *json.RawMessage `json:"Users"`
	}
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &resp)
	if err != nil {
		return nil, ErrRESTServerError
	}
	out := new(SearchResults)
	if err = json.Unmarshal(*resp.Posts, &out.Posts); err != nil {
		out.Posts = make([]*Post, 0)
	}
	if err = json.Unmarshal(*resp.Releases, &out.Releases); err != nil {
		out.Releases = make([]*Release, 0)
	}
	if err = json.Unmarshal(*resp.Comments, &out.Comments); err != nil {
		out.Comments = make([]*Comment, 0)
	}
	if err = json.Unmarshal(*resp.Channels, &out.Channels); err != nil {
		out.Channels = make([]*Channel, 0)
	}
	if err = json.Unmarshal(*resp.Users, &out.Users); err != nil {
		out.Users = make([]*User, 0)
	}
	return out, nil
}
