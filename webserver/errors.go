package webserver

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// errorHandler handles all errors returned by route handlers.
func errorHandler(c *fiber.Ctx, err error) error {
	var r *APIResponse
	var e *fiber.Error
	switch {
	case errors.As(err, &r):
		return c.Status(r.Code).JSON(r)
	case errors.As(err, &e):
		return c.Status(e.Code).JSON(err)
	default:
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
}

// errBadRequest returns a bad request error.
func errBadRequest(msg string) error {
	return newAPIResponse(false, codeBadRequest, msg)
}

// errUnauthorized returns an unauthorized error.
func errUnauthorized(msg string) error {
	return newAPIResponse(false, codeUnauthorized, msg)
}

// errInternal returns a server error.
func errInternal(err error) error {
	return newAPIResponse(false, codeInternal, "Something unexpected happened. Please try again later.")
}
