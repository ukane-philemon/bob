package webserver

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/ukane-philemon/bob/db"
)

const (
	// minUsernameChar is the minimum username character accepted.
	minUsernameChar = 3
	// minPasswordChar is the minimum password character accepted
	minPasswordChar = 8
)

// handleUsernameExists handles the "GET /api/username-exists?username="name""
// endpoint and checks if a username exists.
func (r *WebServer) handleUsernameExists(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" || len(username) < minUsernameChar {
		return errBadRequest(fmt.Sprintf("username with at least %d characters is required", minUsernameChar))
	}

	exists, err := r.db.UsernameExists(username)
	if err != nil {
		return errInternal(err)
	}

	resp := &usernameExitsResponse{
		APIResponse: newAPIResponse(false, codeOk, "Request successful."),
		Exists:      exists,
	}

	return c.Status(resp.Code).JSON(resp)
}

// handleCreateAccount handles the "POST /api/user" endpoint and creates a new
// user account.
func (r *WebServer) handleCreateAccount(c *fiber.Ctx) error {
	form := new(createAccountRequest)
	if err := c.BodyParser(form); err != nil {
		return errBadRequest("invalid request body")
	}

	if form.Username == "" || len(form.Username) < minUsernameChar {
		return errBadRequest(fmt.Sprintf("username with at least %d characters is required", minUsernameChar))
	}

	if !isValidEmail(form.Email) {
		return errBadRequest("a valid email is required")
	}

	if len(form.Password) < 8 {
		return errBadRequest(fmt.Sprintf("password must be a minimum of %d characters", minPasswordChar))
	}
	defer form.Password.Zero()

	if err := r.db.CreateUser(form.Username, form.Email, form.Password.Bytes()); err != nil {
		if errors.Is(err, db.ErrorBadRequest) {
			return errBadRequest(err.Error())
		}
		return errInternal(err)
	}

	resp := newAPIResponse(true, codeOk, "Account created.")

	return c.Status(resp.Code).JSON(resp)
}

// handleGetUser handles the "GET /api/user" endpoint and retrieves a user
// information.
func (r *WebServer) handleGetUser(c *fiber.Ctx) error {
	email, ok := c.Context().UserValue(ctxID).(string)
	if !ok {
		return errInternal(errors.New("Unauthorized user reached an authenticated endpoint"))
	}

	user, err := r.db.RetrieveUserInfo(email)
	if err != nil {
		return errInternal(err)
	}

	resp := &userInfoResponse{
		APIResponse: newAPIResponse(true, codeOk, "User Information Retrieved."),
		Data:        user,
	}

	return c.Status(codeOk).JSON(resp)
}

// handleLogin handles the "POST /api/login" endpoint, verifies the provided
// auth credentials and generates an auth token for the user.
func (r *WebServer) handleLogin(c *fiber.Ctx) error {
	form := new(loginRequest)
	if err := c.BodyParser(form); err != nil {
		return errBadRequest("invalid request body")
	}

	if !isValidEmail(form.Email) {
		return errBadRequest("a valid email is required to login")
	}

	if len(form.Password) == 0 {
		return errBadRequest("password is required")
	}
	defer form.Password.Zero()

	user, err := r.db.LoginUser(form.Email, form.Password.Bytes())
	if err != nil {
		if errors.Is(err, db.ErrorBadRequest) {
			return errBadRequest(err.Error())
		}
		return errInternal(err)
	}

	authToken, err := r.authenticator.generateAuthToken(user.Email, user.Username, jwtAudienceUser, tokenExpiry)
	if err != nil {
		return errInternal(err)
	}

	userInfo := &userInfoResponse{
		APIResponse: newAPIResponse(true, codeOk, "User Information Retrieved."),
		Data:        user,
	}

	return c.Status(userInfo.Code).JSON(loginResponse{userInfo, authToken})
}
