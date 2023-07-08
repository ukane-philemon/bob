package webserver

import (
	"net/http"

	"github.com/mileusna/useragent"
	"github.com/ukane-philemon/bob/db"
)

// APIResponse is the response returned by the API.
type APIResponse struct {
	// Ok is true if the request was successful.
	Ok bool `json:"ok"`
	// Code the HTTP code for this response.
	Code int `json:"code"`
	// Message is the message returned by the API.
	Message string `json:"message"`
}

// Error makes it compatible with the `error` interface.
func (ar *APIResponse) Error() string {
	if ar.Ok || ar.Code == http.StatusOK {
		return ""
	}
	return ar.Message
}

// newAPIResponse creates a new APIResponse.
func newAPIResponse(ok bool, code int, message string) *APIResponse {
	return &APIResponse{
		Ok:      ok,
		Code:    code,
		Message: message,
	}
}

// passwordBytes is a byte slice that can be zeroed after use.
type passwordBytes []byte

// Bytes returns the password bytes.
func (p passwordBytes) Bytes() []byte {
	return []byte(p)
}

// Zero zeroes the password bytes.
func (p passwordBytes) Zero() {
	for i := range p {
		p[i] = 0
	}
}

// createAccountRequest is the request body for the POST /api/create-account
// endpoint.
type createAccountRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginRequest is the request body for the POST /api/login endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// createShortURLRequest is the request body for the "POST /api/url" endpoint
type createShortURLRequest struct {
	LongURL string `json:"longURL"`
	// CustomShortURL is the preferred short URL used instead of generating a
	// new one.
	CustomShortURL string `json:"customShortURL"`
}

// usernameExitsResponse is the response returned by the GET
// /api/username-exists endpoint.
type usernameExitsResponse struct {
	*APIResponse
	Exists bool `json:"exists"`
}

// userInfoResponse is the response returned by the GET /api/user endpoint.
type userInfoResponse struct {
	*APIResponse
	Data *db.UserInfo `json:"data"`
}

type loginResponse struct {
	userInfoResponse
	AuthToken string `json:"authToken"`
}

// shortURLResponse is the response returned by the POST /api/url endpoint and
// the GET /api/url endpoint.
type shortURLResponse struct {
	*APIResponse
	Data interface{} `json:"data"` // *db.ShortURLInfo or []*db.ShortURLInfo
}

// userAgent is a wrapper around useragent.UserAgent.
type userAgent struct {
	useragent.UserAgent
}

// parseUserAgent wraps useragent.Parse.
func parseUserAgent(userAgentInfo string) userAgent {
	return userAgent{useragent.Parse(userAgentInfo)}
}

// DeviceType returns a string representation of the device type.
func (ua userAgent) DeviceType() string {
	switch {
	case ua.Bot:
		return "bot"
	case ua.Desktop:
		return "desktop"
	case ua.Mobile:
		return "mobile"
	case ua.Tablet:
		return "tablet"
	default:
		return "unkown"
	}
}

// updateShortURLRequest is the requesrt body to update a short URL.
type updateShortURLRequest struct {
	LongURL string `json:"longURL"`
	Disable *bool  `json:"disable"`
}
