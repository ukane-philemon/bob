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
func (s *WebServer) handleUsernameExists(c *fiber.Ctx) error {
	username := c.Query("username")
	if username == "" || len(username) < minUsernameChar {
		return errBadRequest(fmt.Sprintf("username with at least %d characters is required", minUsernameChar))
	}

	exists, err := s.db.UsernameExists(username)
	if err != nil {
		appLog.Printf("\nerror checking if username exists: %v\n", err)
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
func (s *WebServer) handleCreateAccount(c *fiber.Ctx) error {
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

	password := passwordBytes(form.Password)
	if len(password) < 8 {
		return errBadRequest(fmt.Sprintf("password must be a minimum of %d characters", minPasswordChar))
	}
	defer password.Zero()

	if err := s.db.CreateUser(form.Username, form.Email, password.Bytes()); err != nil {
		if errors.Is(err, db.ErrorBadRequest) {
			return errBadRequest(err.Error())
		}
		return errInternal(err)
	}

	resp := newAPIResponse(true, codeOk, "Account Created.")

	return c.Status(resp.Code).JSON(resp)
}

// handleGetUser handles the "GET /api/user" endpoint and retrieves a user
// information.
func (s *WebServer) handleGetUser(c *fiber.Ctx) error {
	email, ok := c.Context().UserValue(ctxID).(string)
	if !ok {
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	user, err := s.db.RetrieveUserInfo(email)
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
func (s *WebServer) handleLogin(c *fiber.Ctx) error {
	form := new(loginRequest)
	if err := c.BodyParser(form); err != nil {
		return errBadRequest("invalid request body")
	}

	if !isValidEmail(form.Email) {
		return errBadRequest("a valid email is required to login")
	}

	password := passwordBytes(form.Password)
	if len(password) == 0 {
		return errBadRequest("password is required")
	}
	defer password.Zero()

	user, err := s.db.LoginUser(form.Email, password.Bytes())
	if err != nil {
		if errors.Is(err, db.ErrorBadRequest) {
			return errBadRequest(err.Error())
		}

		appLog.Printf("\nerror logging in user: %v\n", err)
		return errInternal(err)
	}

	authToken, err := s.authenticator.generateAuthToken(user.Email, user.Username, jwtAudienceUser, tokenExpiry)
	if err != nil {
		appLog.Printf("\nerror generating auth token: %v\n", err)
		return errInternal(err)
	}

	userInfo := userInfoResponse{
		APIResponse: newAPIResponse(true, codeOk, "Login Successful."),
		Data:        user,
	}

	return c.Status(userInfo.Code).JSON(loginResponse{userInfo, authToken})
}
