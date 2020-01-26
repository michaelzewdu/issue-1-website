package issue1

import (
	"encoding/json"
	. "fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// ChannelService is used to interact with the Channel Service on the REST server.
type ChannelService service

// SortChannelsBy  holds enums used by to specify the attribute entities are sorted with
type SortChannelsBy string

// FeedSorting constants used by SearchChannels methods
const (
	SortChannelsByCreationTime SortChannelsBy = "creation_time"
	SortChannelsByUsername     SortChannelsBy = "channelUsername"
	SortChannelsByFirstName    SortChannelsBy = "name"
)

// ErrPostAlreadyStickied is returned when the post provided is already a sticky post
var ErrPostAlreadyStickied = Errorf("post already stickied")

// ErrStickiedPostFull is returned when the channel has filled it's stickied post quota
var ErrStickiedPostFull = Errorf("two posts already stickied")

// ErrReleaseAlreadyExists is returned when a release already exists
var ErrReleaseAlreadyExists = Errorf("release already exists")

// ErrAdminAlreadyExists is returned when the channel channelUsername specified already has specified user as admin
var ErrAdminAlreadyExists = Errorf("user is already an admin")

// ErrChannelNotFound is returned when the specified channel does not exist
var ErrChannelNotFound = Errorf("channel does not exist ")

// ErrAdminNotFound is returned when the channel Admin channelUsername specified isn't recognized
var ErrAdminNotFound = Errorf("admin not found")

// ErrStickiedPostNotFound is returned when the  stickied post specified isn't recognized
var ErrStickiedPostNotFound = Errorf("stickied post not found")

// AddChannel sends a a request to create a user based on the passed in struct to the
// REST server. Returns ErrInvalidData if the struct has unacceptable data.
func (s *ChannelService) AddChannel(c *Channel, authToken string) (*Channel, error) {
	var (
		method = http.MethodPost
		path   = Sprintf("/channels")
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
		case http.StatusConflict:
			switch jF.ErrorReason {

			case "channelUsername":
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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}

	c = new(Channel)
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

// GetChannelAuthorized returns the channel under the given channelUsername with private info.

func (s *ChannelService) GetChannelAuthorized(channelUsername string, authToken string) (*Channel, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s", channelUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return nil, err
	}
	addJWTToRequest(req, authToken)

	return s.getChannel(req)
}

// GetChannel returns the channel under the given channelUsername. To get private info of a
// channel, channel GetChannelAuthorized.
func (s *ChannelService) GetChannel(channelUsername string) (*Channel, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s", channelUsername)
	)
	req := s.client.newRequest(path, method)
	return s.getChannel(req)
}

func (s *ChannelService) getChannel(req *http.Request) (*Channel, error) {
	js, statusCode, err := s.client.do(req)
	if err != nil {
		return nil, err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		return nil, ErrChannelNotFound
	case "error":
		return nil, ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}
	c := new(Channel)
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

// GetChannels returns a list of all channels. They are sorted according to the
// default sorting on the REST server. To specify sorting, channel SearchChannels and
// channel an empty string for the pattern.
func (s *ChannelService) GetChannels(page, perPage uint) ([]*Channel, error) {
	p := PaginateParams{}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchChannels("", "", p)
}

// SearchChannelsPaged is a utility wrapper for SearchChannels for easy pagination,
func (s *ChannelService) SearchChannelPaged(page, perPage uint, pattern string, by SortChannelsBy, order SortOrder) ([]*Channel, error) {
	p := PaginateParams{
		SortOrder: order,
	}
	p.Limit, p.Offset = calculateLimitOffset(page, perPage)
	return s.SearchChannels(pattern, by, p)
}

// SearchChannels returns a list of channel according to the passed in parameters.
// An empty pattern matches all channels. If any of the fields on the passed in
// PaginateParams are omitted, it'll use the default values.
func (s *ChannelService) SearchChannels(pattern string, by SortChannelsBy, params PaginateParams) ([]*Channel, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels")
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
			qString = Sprintf("%s_%s", by, params.SortOrder)
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
		return nil, ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return nil, ErrAccessDenied
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}

	channels := make([]*Channel, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &channels)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return channels, nil
}

// UpdateChannel updates the channel under the given channelUsername based on the passed in struct.
// When changing channelUsername, be sure to get new tokens after this call as the one used
// here won't work.
func (s *ChannelService) UpdateChannel(channelUsername string, u *Channel, authToken string) (*Channel, error) {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s", channelUsername)
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
			return nil, ErrChannelNotFound
		case http.StatusConflict:
			switch jF.ErrorReason {

			case "channelUsername":
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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}

	c := new(Channel)
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

// DeleteUser removes the user under the given channelUsername.
func (s *ChannelService) DeleteChannel(channelUsername, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s", channelUsername)
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
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// AddPicture sets the passed in image as the channel's picture for the channel
// under the passed in channelUsername.
func (s *ChannelService) AddPicture(channelUsername string, image io.Reader, imageName, authToken string) (string, error) {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s/picture", channelUsername)
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
			case "channelUsername":
				return "", ErrPostNotFound
			}
		case http.StatusNotFound:
			return "", ErrChannelNotFound
		default:
			return "", ErrRESTServerError
		}
	case "error":
		return "", ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return "", ErrAccessDenied
		case http.StatusForbidden:
			return "", ErrForbiddenAccess
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

// RemovePicture picture removes the picture of the channel under the given
// channelUsername.
func (s *ChannelService) RemovePicture(channelUsername, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s/picture", channelUsername)
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
		case http.StatusForbidden:
			return ErrForbiddenAccess
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// AddAdmin adds an admin into the channel Admin list under the given channelUsername based on the passed in admin channelUsername.

func (s *ChannelService) AddAdmin(channelUsername string, adminUsername string, authToken string) error {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s/admins/%s", channelUsername, adminUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusConflict:
			return ErrAdminAlreadyExists
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "adminUsername":
				return ErrAdminNotFound
			default:
			}
			fallthrough

		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// DeleteAdmin deletes the channel Admin list under the given channelUsername based on the passed in admin Username.
// When deleting admin, be sure to get new tokens after this call as the one used
// here might won't work.
func (s *ChannelService) DeleteAdmin(channelUsername string, adminUsername string, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s/admins/%s", channelUsername, adminUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case adminUsername:
				return ErrAdminNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// DeleteAdmin deletes the channel Admin list under the given channelUsername based on the passed in admin Username.
// When deleting admin, be sure to get new tokens after this call as the one used
// here might won't work.
func (s *ChannelService) ChangeOwner(channelUsername string, ownerUsername string, authToken string) error {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s/owners/%s", channelUsername, ownerUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "ownerUsername":
				return ErrAdminNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// DeleteReleaseFromCatalog deletes the channel catalog list under the given channelUsername based on the passed in ReleaseId.

func (s *ChannelService) DeleteReleaseFromCatalog(channelUsername string, releaseID uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s/catalogs/%d", channelUsername, releaseID)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "releaseID":
				return ErrReleaseNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

// DeleteReleaseFromOfficialCatalog deletes the channel official catalog list under the given channelUsername based on the passed in ReleaseId.
func (s *ChannelService) DeleteReleaseFromOfficialCatalog(channelUsername string, releaseID uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s/official/%d", channelUsername, releaseID)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "releaseID":
				return ErrReleaseNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

//Add Release to Official Catalog adds a release from the channel catalog to the official catalog
func (s *ChannelService) AddReleaseToOfficialCatalog(channelUsername string, releaseID int, postID uint, authToken string) error {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s/official/%d", channelUsername, releaseID)
	)
	req := s.client.newRequest(path, method)

	var requestData struct {
		PostID uint `json:"postID"` //postFrom ID
	}
	requestData.PostID = postID
	err := addBodyToRequestAsJSON(req, requestData)
	if err != nil {
		return err
	}
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
		case http.StatusConflict:
			if jF.ErrorReason == "releaseID" {
				return ErrReleaseAlreadyExists
			}

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "releaseID":
				return ErrReleaseNotFound
			case "postID":
				return ErrPostNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

//Sticky Post stickies two posts on top of channel post view
func (s *ChannelService) StickyPost(channelUsername string, postID uint, authToken string) error {
	var (
		method = http.MethodPut
		path   = Sprintf("/channels/%s/Posts/%d", channelUsername, postID)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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
		case http.StatusServiceUnavailable:
			switch jF.ErrorReason {

			case "Stickied postID":
				return ErrStickiedPostFull
			default:
			}
			fallthrough
		case http.StatusConflict:
			switch jF.ErrorReason {

			case "stickiedPostID":
				return ErrPostAlreadyStickied
			default:
			}
			fallthrough
		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "postID":
				return ErrPostNotFound
			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

//Sticky Post stickies two posts on top of channel post view
func (s *ChannelService) DeleteStickiedPost(channelUsername string, postID uint, authToken string) error {
	var (
		method = http.MethodDelete
		path   = Sprintf("/channels/%s/stickiedPosts/%d", channelUsername, postID)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return err
	}
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

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return ErrChannelNotFound
			case "stickiedPostID":
				return ErrStickiedPostNotFound

			default:
			}
			fallthrough
		default:
			return ErrRESTServerError
		}
	case "error":
		return ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return ErrAccessDenied
		case http.StatusForbidden:
			return ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return ErrRESTServerError
		}
	}
	return nil
}

//// GetChannelPosts returns the channel post under the given channelUsername.
func (s *ChannelService) GetChannelPosts(channelUsername string) ([]*Post, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/Posts", channelUsername)
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

			case "channelUsername":
				return nil, ErrChannelNotFound
			case "postID":
				return nil, ErrPostNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
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

//TODO
//// GetCatalog returns the channel catalog under the given channelUsername.
func (s *ChannelService) GetCatalog(channelUsername string, authToken string) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/catalog", channelUsername)
	)

	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
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

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return nil, ErrChannelNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
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
	err = json.Unmarshal(*data, releases)

	if err != nil {
		return nil, ErrRESTServerError
	}

	return releases, nil
}

//// GetOfficialCatalog returns the channel Official catalog under the given channelUsername.
func (s *ChannelService) GetOfficialCatalog(channelUsername string, authToken string) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/official", channelUsername)
	)

	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
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

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return nil, ErrChannelNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
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
	err = json.Unmarshal(*data, releases)

	if err != nil {
		return nil, ErrRESTServerError
	}

	return releases, nil
}

// GetChannelPosts returns the channel post under the given channelUsername.
func (s *ChannelService) GetChannelPost(channelUsername string, postId uint) (*Post, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/Posts/%d", channelUsername, postId)
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

			case "channelUsername":
				return nil, ErrChannelNotFound
			case "postID":
				return nil, ErrPostNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
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

// GetChannelPosts returns the channel post under the given channelUsername.
func (s *ChannelService) GetReleaseInCatalog(channelUsername string, releaseId uint, authToken string) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/catalogs/%d", channelUsername, releaseId)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
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

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return nil, ErrChannelNotFound
			case "releaseID":
				return nil, ErrReleaseNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}
	r := make([]*Release, 0)
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &r)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return r, nil
}

// GetReleaseInOfficialCatalog returns the channel releases in official under the given channelUsername.
func (s *ChannelService) GetReleaseInOfficialCatalog(channelUsername string, releaseId uint) ([]*Release, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/official/%d", channelUsername, releaseId)
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

			case "channelUsername":
				return nil, ErrChannelNotFound
			case "releaseID":
				return nil, ErrReleaseNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
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

// GetChannelPosts returns the channel post under the given channelUsername.
func (s *ChannelService) GetStickiedPosts(channelUsername string) ([]*Post, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/stickiedPosts", channelUsername)
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

			case "channelUsername":
				return nil, ErrChannelNotFound
			case "stickiedPostID":
				return nil, ErrPostNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}
	posts := make([]*Post, 0)
	if &posts == nil {
		return nil, ErrPostNotFound
	}
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	if *data == nil {
		return nil, ErrPostNotFound
	}
	err = json.Unmarshal(*data, &posts)
	if err != nil {
		Printf("%s", err.Error())
		return nil, ErrRESTServerError
	}

	return posts, nil
}

// GetAdmins returns the channel admins under the given channelUsername.
func (s *ChannelService) GetAdmins(channelUsername string, authToken string) ([]string, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/admins", channelUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
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

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return nil, ErrChannelNotFound

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
		case http.StatusForbidden:
			return nil, ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:
			return nil, ErrRESTServerError
		}
	}
	var admins []string
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return nil, ErrRESTServerError
	}
	err = json.Unmarshal(*data, &admins)
	if err != nil {
		return nil, ErrRESTServerError
	}
	return admins, nil
}

// GetAdmins returns the channel admins under the given channelUsername.
func (s *ChannelService) GetOwner(channelUsername string, authToken string) (string, error) {
	var (
		method = http.MethodGet
		path   = Sprintf("/channels/%s/owners", channelUsername)
	)
	req := s.client.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, authToken)
	if err != nil {
		return "", err
	}
	addJWTToRequest(req, authToken)

	js, statusCode, err := s.client.do(req)
	if err != nil {
		return "", err
	}

	switch js.Status {
	case "success":
		break
	case "fail":
		jF, ok := js.Data.(*jSendFailData)
		if !ok {
			s.client.Logger.Printf("tHIS0")
			return "", ErrRESTServerError
		}
		s.client.Logger.Printf("%+v", jF)
		switch statusCode {
		case http.StatusBadRequest:
			return "", ErrInvalidData

		case http.StatusNotFound:
			switch jF.ErrorReason {

			case "channelUsername":
				return "", ErrChannelNotFound

			default:
			}
			fallthrough
		default:
			return "", ErrRESTServerError
		}
	case "error":
		return "", ErrRESTServerError
	default:
		switch statusCode {
		case http.StatusUnauthorized:
			return "", ErrAccessDenied
		case http.StatusForbidden:
			return "", ErrForbiddenAccess
		case http.StatusInternalServerError:
			fallthrough
		default:

			return "", ErrRESTServerError
		}
	}
	var owner string
	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		s.client.Logger.Printf("tH4IS")
		return "", ErrRESTServerError
	}
	err = json.Unmarshal(*data, &owner)
	if err != nil {

		return "", ErrRESTServerError
	}
	return owner, nil
}
