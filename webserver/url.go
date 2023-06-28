package webserver

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v2"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/ukane-philemon/bob/db"
)

// handleCreateShortURL handles the "POST /api/url" endpoint and creates a new
// short URL.
func (s *WebServer) handleCreateShortURL(c *fiber.Ctx) error {
	form := new(createShortURLRequest)
	if err := c.BodyParser(form); err != nil {
		return errBadRequest("invalid request body")
	}

	url, err := url.ParseRequestURI(form.URL)
	if err != nil || url.Scheme != "https" || url.Host == "" || url.Path == "" {
		return errBadRequest("invalid URL, provide an absolute URL with a scheme (only https is allowed) and a host (e.g. https://example.com/path/to/resource))")
	}

	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	req := a.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(url.String())

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	if err = a.Parse(); err != nil {
		return errInternal(err)
	}

	// Ensure the url is reachable. We allow a maximum of 2 redirects.
	if err := a.DoRedirects(req, resp, 2); err != nil {
		return errBadRequest(err.Error())
	}

	if resp.StatusCode() != fiber.StatusOK {
		return errBadRequest("invalid URL, the URL is not reachable")
	}

	apiResp := &shortURLResponse{
		APIResponse: newAPIResponse(true, codeOk, "URL created successfully"),
	}

	if email, ok := c.Context().UserValue(ctxID).(string); ok {
		apiResp.Data, err = s.db.SaveUserURL(email, form.URL)
	} else if id := c.IP(); id != "" {
		apiResp.Data, err = s.db.SaveGuestURL(id, form.URL)
	} else {
		return errBadRequest("invalid request")
	}
	if err != nil {
		return translateDBError(err)
	}

	return c.Status(codeOk).JSON(apiResp)
}

// handleGetAllURL handles the "GET /api/url" endpoint and returns all the short
// URLs for a validated user.
func (s *WebServer) handleGetAllURL(c *fiber.Ctx) error {
	email, ok := c.Context().UserValue(ctxID).(string)
	if !ok { // not logged in but we should never reach here
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	urls, err := s.db.RetrieveUserURLs(email)
	if err != nil {
		return translateDBError(err)
	}

	apiResp := &shortURLResponse{
		APIResponse: newAPIResponse(true, codeOk, "URLs retrieved successfully"),
		Data:        urls,
	}

	return c.Status(codeOk).JSON(apiResp)
}

// handleGetURL handles the "GET /url/{shortUrl} "endpoint and returns the full
// information about a short URL for a validated user.
func (s *WebServer) handleGetURL(c *fiber.Ctx) error {
	if _, ok := c.Context().UserValue(ctxID).(string); !ok { // not logged in but we should never reach here
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortUrl := c.Params("shortUrl")
	if shortUrl == "" || len(shortUrl) > db.URLLength+10 /* give a benefit of doubt */ {
		return errBadRequest("invalid short URL")
	}

	urlInfo, err := s.db.RetrieveURLInfo(shortUrl)
	if err != nil {
		return translateDBError(err)
	}

	apiResp := &shortURLResponse{
		APIResponse: newAPIResponse(true, codeOk, "URL retrieved successfully"),
		Data:        urlInfo,
	}

	return c.Status(codeOk).JSON(apiResp)
}

// handleCreateURLQR handles the "GET /api/url/{shortUrl}/qr" endpoint and returns
// a QR code for the short URL.
func (s *WebServer) handleCreateURLQR(c *fiber.Ctx) error {
	if _, ok := c.Context().UserValue(ctxID).(string); !ok { // not logged in but we should never reach here
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortUrl := c.Params("shortUrl")
	if shortUrl == "" || len(shortUrl) > db.URLLength+10 /* give a benefit of doubt */ {
		return errBadRequest("invalid short URL")
	}

	urlInfo, err := s.db.RetrieveURLInfo(shortUrl)
	if err != nil {
		return translateDBError(err)
	}

	fullUrl := fmt.Sprintf("%s/%s", s.addr, urlInfo.ShortURL)
	png, err := qrcode.Encode(fullUrl, qrcode.Medium, 200)
	if err != nil {
		return errBadRequest("failed to generate QR code")
	}

	c.Set("Content-Type", "image/png")
	c.Set("Content-Length", strconv.Itoa(len(png)))
	c.Status(codeOk)
	return c.Send(png)
}

// handleShortUrlRedirect handles the "GET /{shortUrl}" endpoint and redirects to
// the original URL.
func (s *WebServer) handleShortUrlRedirect(c *fiber.Ctx) error {
	shortUrl := c.Params("shortUrl")
	if shortUrl == "" || len(shortUrl) > db.URLLength+10 /* give a benefit of doubt */ {
		return errBadRequest("invalid short URL")
	}

	urlInfo, err := s.db.RetrieveURLInfo(shortUrl)
	if err != nil {
		return translateDBError(err)
	}

	// Update the short URL stats in the background. TODO: Log the error if any.
	defer s.db.UpdateShortURL(shortUrl)

	return c.Redirect(urlInfo.OriginalURL, codeFound)
}
