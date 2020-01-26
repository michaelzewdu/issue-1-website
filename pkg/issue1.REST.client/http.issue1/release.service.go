package issue1

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ReleaseService is used to interact with the release service on the REST server.
type ReleaseService service

// SortReleasesBy  holds enums used by SearchRelease methods the attribute of Users are sorted with
type SortReleasesBy string

// Sorting constants used by SearchRelease methods
const (
	SortReleaseByCreationTime SortReleasesBy = "creation-time"
	SortByChannel             SortReleasesBy = "channel"
	SortByType                SortReleasesBy = "type"
)

// GetRelease returns the user under the given username. To be able to get an unofficial
// release use GetReleaseAuthorized.
func (s *ReleaseService) GetRelease(id uint) (*Release, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/releases/%d", id)
	)
	req := s.client.newRequest(path, method)
	return s.getRelease(req)
}

// GetReleaseAuthorized retrieves releases and possibly unofficial releases from channels
// the user the auth token is provided for is an admin of.
func (s *ReleaseService) GetReleaseAuthorized(id uint, token string) (*Release, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/releases/%d", id)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, token)
	return s.getRelease(req)
}

func (s *ReleaseService) getRelease(req *http.Request) (*Release, error) {
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		return nil, ErrReleaseNotFound
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
	r := new(Release)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, r)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return r, nil
}

// AddTextRelease sends a request to add a text release based on the given struct. AuthToken
// of an admin of the channel the release is being added to must be passed as well.
func (s *ReleaseService) AddTextRelease(r *Release, authToken string) (*Release, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("/releases")
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	r.Type = Text
	err := addBodyToRequestAsJSON(req, r)
	if err != nil {
		return nil, err
	}
	return s.addRelease(req)
}

// AddImageRelease sends a request to add an image release based on the given struct. AuthToken
// of an admin of the channel the release is being added to must be passed as well.
func (s *ReleaseService) AddImageRelease(r *Release, image io.Reader, imageName, authToken string) (*Release, error) {
	var (
		method = http.MethodPost
		path   = fmt.Sprintf("/releases")
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	r.Type = Image
	err := addJSONAndImageToRequestAsMultipart(req, r, image, imageName)
	if err != nil {
		return nil, err
	}

	return s.addRelease(req)
}

func (s *ReleaseService) addRelease(req *http.Request) (*Release, error) {
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
			return nil, ErrChannelNotFound
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "image-type":
				return nil, ErrUnacceptedImageType
			case "image":
				return nil, ErrInvalidData
			default:
				return nil, ErrInvalidData
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

	rel := new(Release)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, rel)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return rel, nil
}

// UpdateTextRelease sends a request to update release based on the given struct. Release type must
// be passed in to resolve ambiguities of intentions. Use UpdateImageRelease to update the images of
// image based releases AuthToken of an admin of the channel the release is being added to must be
// passed as well.
func (s *ReleaseService) UpdateRelease(id uint, r *Release, t ReleaseType, authToken string) (*Release, error) {
	var (
		method = http.MethodPatch
		path   = fmt.Sprintf("/releases/%d", id)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	r.Type = t
	err := addBodyToRequestAsJSON(req, r)
	if err != nil {
		return nil, err
	}
	return s.updateRelease(req)
}

// UpdateImageRelease sends a request to update an image release based on the given struct. AuthToken
// of an admin of the channel the release is being added to must be passed as well.
func (s *ReleaseService) UpdateImageRelease(id uint, r *Release, image io.Reader, imageName, authToken string) (*Release, error) {
	var (
		method = http.MethodPatch
		path   = fmt.Sprintf("/releases/%d", id)
	)
	req := s.client.newRequest(path, method)
	addJWTToRequest(req, authToken)
	r.Type = Image
	err := addJSONAndImageToRequestAsMultipart(req, r, image, imageName)
	if err != nil {
		return nil, err
	}

	return s.updateRelease(req)
}

func (s *ReleaseService) updateRelease(req *http.Request) (*Release, error) {
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
			return nil, ErrChannelNotFound
		case http.StatusBadRequest:
			switch jF.ErrorReason {
			case "image-type":
				return nil, ErrUnacceptedImageType
			case "image":
				return nil, ErrInvalidData
			default:
				return nil, ErrInvalidData
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

	rel := new(Release)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, rel)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return rel, nil
}

// DeleteRelease removes the release under the given id.
func (s *ReleaseService) DeleteRelease(id uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = fmt.Sprintf("/releases/%d", id)
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

// GetReleases returns a list of all releases. They are sorted according to the
// default sorting on the REST server. To specify sorting, user SearchReleases and
// user an empty string for the pattern.
// Note: You can only search for releases found in official catalogs of channels.
func (s *ReleaseService) GetReleases(page, perPage uint) ([]*Release, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchReleases("", "", p)
}

// SearchReleasesPaged is a utility wrapper for SearchReleases for easy pagination,
// Note: You can only search for releases found in official catalogs of channels.
func (s *ReleaseService) SearchReleasesPaged(page, perPage uint, pattern string, by SortReleasesBy, order SortOrder) ([]*Release, error) {
	p := PaginateParams{
		SortOrder: order,
	}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchReleases(pattern, by, p)
}

// SearchReleases returns a list of releases according to the passed in parameters.
// An empty pattern matches all releases. If any of the fields on the passed in
// PaginateParams are omitted, it'll use the default values.
// Note: You can only search for releases found in official catalogs of channels.
func (s *ReleaseService) SearchReleases(pattern string, by SortReleasesBy, params PaginateParams) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = fmt.Sprintf("/releases")
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
