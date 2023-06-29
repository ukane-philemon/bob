package webserver

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// validateIfLoggedIn is a middleware handle that validates a user token, if
// present. Each endpoint will reject the request if use login is required.
func (s *WebServer) validateIfLoggedIn(c *fiber.Ctx) error {
	authHeader := c.Get(fiber.HeaderAuthorization)
	if authHeader == "" {
		return c.Next() // No auth header, so no user is logged in.
	}

	authTokenParts := strings.Split(authHeader, " ")
	if len(authTokenParts) != 2 || !strings.EqualFold(authTokenParts[0], "Bearer") || authTokenParts[1] == "" {
		return errUnauthorized("Invalid authorization header")
	}

	// Get the auth token and validated it.
	token, ok := s.authenticator.validateAuthToken(strings.TrimSpace(authTokenParts[1]))
	if !ok {
		return errUnauthorized("Invalid authorization token")
	}

	// Set the user email in the context.
	c.Context().SetUserValue(ctxID, token.ID)
	return c.Next()
}
