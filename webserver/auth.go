package webserver

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// requireUserLogin is a middleware handler that ensures that a user is logged
// in before access to specific endpoints are allowed.
func (s *WebServer) requireUserLogin(c *fiber.Ctx) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return errUnauthorized("Missing authorization header")
	}

	authTokenParts := strings.Split(authHeader, " ")
	if len(authTokenParts) != 2 || !strings.EqualFold(authTokenParts[0], "Bearer") || authTokenParts[1] == "" {
		return errUnauthorized("Invalid authorization header")
	}

	// Get the auth token and validated it.
	token, ok := s.authenticator.validateAuthToken(authTokenParts[1])
	if !ok {
		return errUnauthorized("Invalid authorization token")
	}

	// Set the user email in the context.
	c.Context().SetUserValue(ctxID, token.ID)
	return nil
}

// validateIfLoggedIn is a middleware handle that validates a user token, if
// present. This is used for endpoints that allows guest and logged in users.
func (s *WebServer) validateIfLoggedIn(c *fiber.Ctx) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return nil // No auth header, so no user is logged in.
	}

	return s.requireUserLogin(c)
}
