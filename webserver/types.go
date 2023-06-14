package webserver

import (
	"net/http"

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
	Username string        `json:"username"`
	Email    string        `json:"email"`
	Password passwordBytes `json:"password"`
}

// loginRequest is the request body for the POST /api/login endpoint.
type loginRequest struct {
	Email    string        `json:"email"`
	Password passwordBytes `json:"password"`
}

// createShortURLRequest is the request body for the "POST /api/url" endpoint
type createShortURLRequest struct {
	URL string `json:"url"`
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
	Data *db.User `json:"data"`
}

type loginResponse struct {
	*userInfoResponse
	AuthToken string `json:"authToken"`
}

// shortURLResponse is the response returned by the POST /api/url endpoint and
// the GET /api/url endpoint.
type shortURLResponse struct {
	*APIResponse
	Data interface{} `json:"data"` // *db.ShortURLInfo or []*db.ShortURLInfo
}
