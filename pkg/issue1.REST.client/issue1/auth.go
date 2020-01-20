package issue1

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAuthToken gets an JWT auth token using the provided credentials.
func (c *Client) GetAuthToken(username, password string) (string, error) {
	var (
		path   = fmt.Sprintf("/token-auth")
		method = http.MethodPost
	)
	req := c.newRequest(path, method)

	err := addBodyToRequestAsJSON(req, struct {
		Username string `json:"username"`
		Password string `json:"password,omitempty"`
	}{username, password})
	if err != nil {
		return "", err
	}

	js, statusCode, err := c.do(req)
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
		switch statusCode {
		case http.StatusUnauthorized:
			return "", ErrCredentialsUnaccepted
		case http.StatusBadRequest:
			fallthrough
		default:
			return "", ErrRESTServerError
		}
	case "error":
		fallthrough
	default:
		return "", ErrRESTServerError
	}

	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return "", ErrRESTServerError
	}

	var t struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(*data, &t)
	if err != nil {
		return "", ErrRESTServerError
	}
	return t.Token, nil
}

// RefreshAuthToken gets a new token using the passed in token.
// If the passed in token is too old, it will throw ErrAccessDenied.
func (c *Client) RefreshAuthToken(token string) (string, error) {
	var (
		path   = fmt.Sprintf("/token-auth-refresh")
		method = http.MethodGet
	)
	req := c.newRequest(path, method)
	addJWTToRequest(req, token)

	js, statusCode, err := c.do(req)
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
		switch statusCode {
		case http.StatusBadRequest:
			fallthrough
		default:
			return "", ErrRESTServerError
		}
	case "error":
		fallthrough
	default:
		return "", ErrRESTServerError
	}

	data, ok := js.Data.(*json.RawMessage)
	if !ok {
		return "", ErrRESTServerError
	}

	var t struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(*data, &t)
	if err != nil {
		return "", ErrRESTServerError
	}
	return t.Token, nil
}

// Logout invalidates the passed in token from further usage.
func (c *Client) Logout(token string) error {
	var (
		path   = fmt.Sprintf("/logout")
		method = http.MethodGet
	)
	req := c.newRequest(path, method)
	addJWTToRequest(req, token)

	js, statusCode, err := c.do(req)
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
		case http.StatusBadRequest:
			fallthrough
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
