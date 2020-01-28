/*
Package issue1 provides provides an interface to the issue#1 REST services.*/
package issue1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
)

/*
type BadDataError struct {
	Field string
	Message string
}

func (e BadDataError) Error() string {
	return fmt.Sprintf("field: %s message:%s", e.Field, e.Message)
}
*/

// these errors shouldn't be used outside this package
var (
	//errUnacceptableData          = errors.New("request made by this returned 400: possible client out of date")
	//errRestServerError           = errors.New("response was 500: contact the dev")
	//errNotFound                  = errors.New("404: requested resource not found")
	//errUnauthorized              = errors.New("401: unauthorized")
	errJSONDeserializationFailed = errors.New("was unable to deserialize JSON in response")
	//errUnrecognizedResponse      = errors.New("unrecognized response status code")
	//errStatusConflict            = errors.New("409: possible occupied id")
)

var (
	// ErrRESTServerError is usually returned when the response from the REST
	// servers is unexpected and un-parse-able. This usually means a change in protocol
	// or an error in this client.
	ErrRESTServerError = errors.New("http.issue1: rest server error")
	// ErrAccessDenied is returned when server returns a 401:unauthorized either because
	// the token was unaccepted or because the token doesn't give access to the resource.
	ErrAccessDenied = errors.New("http.issue1: access denied")
	//ErrCredentialsUnaccepted is returned if the given username:password combo is wrong.
	ErrCredentialsUnaccepted = errors.New("http.issue1: credentials not accepted")
	//ErrConnectionError is return if there was an error sending a request to the REST server.
	ErrConnectionError = errors.New("http.issue1: connection could not be made with issue1 REST")
	//ErrForbiddenAccess is return if there was a forbidden request is sent to the REST server.
	ErrForbiddenAccess = errors.New("http.issue1:forbidden URL request")
	//ErrInvalidData is usually returned when the passed data is missing required fields or
	// is malformed.
	ErrInvalidData = errors.New("http.issue1: provided data was not accepted")
	//ErrUserNotFound is returned when there's no user found under the passed in username.
	ErrUserNotFound = errors.New("http.issue1: user was not found")
	//ErrPostNotFound is returned when there was no post found under the given id.
	ErrPostNotFound = errors.New("http.issue1: post was not found")
	//ErrReleaseNotFound is returned when there was no post found under the given id.
	ErrReleaseNotFound = errors.New("http.issue1: release was not found")
	//ErrCommentNotFound is returned when there was no comment found under the given id.
	ErrCommentNotFound = errors.New("http.issue1: comment was not found")
	// ErrUnacceptedImageType is returned when the image format passed isn't supported by REST.
	ErrUnacceptedImageType = errors.New("http.issue1: file mime type not accepted")

	//ErrStarNotFound is returned when requested Star is not found
	ErrStarNotFound = fmt.Errorf("Specified star not found")
)

// SortOrder holds enums used to specify the order entities are sorted with
type SortOrder string

const (
	// SortAscending sorts accordingly.
	SortAscending SortOrder = "asc"
	// SortDescending sorts in descending manner.
	SortDescending SortOrder = "dsc"
)

// PaginateParams is used to pass in parameters to the Search methods of issue1 services.
type PaginateParams struct {
	SortOrder SortOrder
	Limit     uint
	Offset    uint
}

type service struct {
	client *Client
}

// Client is a type used to interact with the issue1 REST servers.
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL

	Logger         *log.Logger
	ChannelService ChannelService
	UserService    UserService
	FeedService    FeedService
	PostService    PostService
	ReleaseService ReleaseService
	CommentService CommentService
	SearchService  SearchService
	AuthService
}

// NewClient returns a new issue1 client.
func NewClient(httpClient *http.Client, baseURL *url.URL, logger *log.Logger) *Client {
	c := &Client{HTTPClient: httpClient,
		BaseURL: baseURL,
		Logger:  logger,
	}
	c.UserService = UserService{client: c}
	c.ChannelService = ChannelService{client: c}
	c.FeedService = FeedService{client: c}
	c.AuthService = AuthService{client: c}
	c.ReleaseService = ReleaseService{client: c}
	c.CommentService = CommentService{client: c}
	c.PostService = PostService{client: c}
	c.SearchService = SearchService{client: c}
	return c
}

func (c *Client) newRequest(path string, method string) *http.Request {
	req := &http.Request{
		Method: method,
		URL:    c.BaseURL.ResolveReference(&url.URL{Path: path}),
		Header: make(http.Header),
	}
	return req
}

func (c *Client) newRequestFromURL(u *url.URL, method string) *http.Request {
	req := &http.Request{
		Method: method,
		URL:    c.BaseURL.ResolveReference(u),
		Header: make(http.Header),
	}
	return req
}

func addBodyToRequestAsJSON(req *http.Request, body interface{}) error {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(body)
	if err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(buf)
	req.ContentLength = int64(buf.Len())
	req.Header.Add("Content-Type", "application/json")
	return nil
}

func addImageToRequest(req *http.Request, image io.Reader, imageName string) error {
	buf := new(bytes.Buffer)
	//if req.Body != nil {
	//	_, err := io.Copy(buf, req.Body)
	//	if err != nil {
	//		return err
	//	}
	//}
	mw := multipart.NewWriter(buf)
	defer mw.Close()
	fw, err := mw.CreateFormFile("image", imageName)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, image)
	if err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return nil
}

func addJSONAndImageToRequestAsMultipart(req *http.Request, body interface{}, image io.Reader, imageName string) error {
	buf := new(bytes.Buffer)
	//if req.Body != nil {
	//	_, err := io.Copy(buf, req.Body)
	//	if err != nil {
	//		return err
	//	}
	//}
	mw := multipart.NewWriter(buf)
	defer mw.Close()

	jsonHeader := make(textproto.MIMEHeader)
	jsonHeader.Set("Content-Type", "application/json")
	jsonHeader.Set("Content-Disposition", `form-data; name="JSON"`)

	jw, err := mw.CreatePart(jsonHeader)
	if err != nil {
		return err
	}
	err = json.NewEncoder(jw).Encode(body)
	if err != nil {
		return err
	}

	fw, err := mw.CreateFormFile("image", imageName)
	if err != nil {
		return err
	}
	_, err = io.Copy(fw, image)
	if err != nil {
		return err
	}

	req.Body = ioutil.NopCloser(buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return nil
}

func addJWTToRequest(req *http.Request, token string) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
}

func calculateLimitOffset(page, perPage uint) (limit, offset uint) {
	return perPage, (page - 1) * perPage
}

/*
func (c *Client) do(req *http.Request) (*jSendResponse, error) {
	var err error
	jSend := new(jSendResponse)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusCreated:
	case http.StatusNotFound:
		err = errNotFound
	case http.StatusConflict:
		err = errStatusConflict
	case http.StatusBadRequest:
		err = errUnacceptableData
	case http.StatusUnauthorized:
		err = errUnauthorized
	case http.StatusInternalServerError:
		err = errRestServerError
	default:
		err = errUnrecognizedResponse
	}

	err2 := json.NewDecoder(resp.Body).Decode(jSend)
	if err2 != nil {
		return nil, errJSONDeserializationFailed
	}

	return jSend, nil
}
*/

func (c *Client) do(req *http.Request) (*jSendResponse, int, error) {
	var err error
	jSend := new(jSendResponse)
	//c.Logger.Printf("request: %s %s\n", req.Method, req.URL.String())
	/*
		{
			err = req.Clone(req.Context()).Write(os.Stdout)
			if err != nil {
				c.Logger.Printf("err: %v\n", err)
			}
		}*/
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		if _, ok := err.(net.Error); ok {
			return nil, -1, ErrConnectionError
		}
		return nil, -1, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, resp.StatusCode, ErrAccessDenied
	}

	//buf := new(bytes.Buffer)
	//_, err = io.Copy(buf, resp.Body)
	//fmt.Printf("Body:\n %s", string(buf.Bytes()))
	//err = json.NewDecoder(buf).Decode(jSend)

	err = json.NewDecoder(resp.Body).Decode(jSend)
	if err != nil {
		return nil, -1, ErrRESTServerError
	}
	//c.Logger.Printf("statusCode: %d\n", resp.StatusCode)
	//c.Logger.Printf("response: %+v\n", jSend)
	return jSend, resp.StatusCode, nil
}

type jSendResponse struct {
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type jSendFailData struct {
	ErrorReason  string `json:"errorReason"`
	ErrorMessage string `json:"errorMessage"`
}

// UnmarshalJSON is used to unmarshal responses from the REST servers into
// the jSend format.
func (j *jSendResponse) UnmarshalJSON(b []byte) error {
	// we must use a type different from jSendResponse for the
	// initial Unmarshal to avoid recursive calls ad infinitum
	var doppelganger struct {
		Status  string           `json:"status"`
		Data    *json.RawMessage `json:"data,omitempty"`
		Message string           `json:"message,omitempty"`
	}
	err := json.Unmarshal(b, &doppelganger)
	if err != nil {
		return errJSONDeserializationFailed
	}
	j.Status = doppelganger.Status
	j.Message = doppelganger.Message
	switch j.Status {
	case "success":
		// if successful, they'll have to unmarshal the RawMessage
		// itself to the type they want.
		j.Data = doppelganger.Data
	case "fail":
		failStruct := new(jSendFailData)
		err := json.Unmarshal(*doppelganger.Data, failStruct)
		if err != nil {
			return errJSONDeserializationFailed
		}
		j.Data = failStruct
	case "error":
	default:
	}
	return nil
}
