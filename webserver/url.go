package webserver

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

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

	longURL, err := url.ParseRequestURI(form.LongURL)
	if err != nil || longURL.Scheme != "https" || longURL.Host == "" {
		return errBadRequest("invalid URL, provide an absolute URL with a scheme (only https is allowed) and a host (e.g. https://example.com/path/to/resource))")
	}

	var userID string
	var ok bool
	userID, ok = c.Context().UserValue(ctxID).(string)
	if !ok {
		if form.CustomShortURL != "" {
			return errBadRequest("Create an account to use custom short URL feature")
		}

		userID = c.IP()
	}

	if userID == "" {
		return errBadRequest("invalid request")
	}

	if form.CustomShortURL != "" && !customURLRegEx.MatchString(form.CustomShortURL) {
		return errBadRequest("invalid custom short url")
	}

	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	req := a.Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI(longURL.String())

	resp := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(resp)

	if err = a.Parse(); err != nil {
		appLog.Printf("\nerror parsing URL: %v\n", err)
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
		APIResponse: newAPIResponse(true, codeOk, "Request was successful"),
	}

	url, err := s.db.CreateNewShortURL(userID, form.LongURL, form.CustomShortURL, !isValidEmail(userID))
	if err != nil {
		return translateDBError(err)
	}

	apiResp.Data = url
	s.urlMtx.Lock()
	s.urlCache[url.ShortURL] = url
	s.urlMtx.Unlock()

	return c.Status(codeOk).JSON(apiResp)
}

// handleGetAllURL handles the "GET /api/url" endpoint and returns all the short
// URLs for a validated user.
func (s *WebServer) handleGetAllURL(c *fiber.Ctx) error {
	email, ok := c.Context().UserValue(ctxID).(string)
	if !ok {
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
	if _, ok := c.Context().UserValue(ctxID).(string); !ok {
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortUrl := c.Params("shortUrl")
	if shortUrl == "" {
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
	if _, ok := c.Context().UserValue(ctxID).(string); !ok {
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortUrl := c.Params("shortUrl")
	if shortUrl == "" {
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
	if shortUrl == "" {
		return errBadRequest("invalid short URL")
	}

	s.urlMtx.RLock()
	urlInfo, found := s.urlCache[shortUrl]
	s.urlMtx.RUnlock()
	if !found {
		var err error
		urlInfo, err = s.db.RetrieveURLInfo(shortUrl)
		if err != nil {
			return translateDBError(err)
		}
	}

	if urlInfo.Disabled {
		return errBadRequest("Link has been disabled")
	}

	userAgentBytes := c.Context().UserAgent()
	ua := parseUserAgent(string(userAgentBytes))
	// Update the short URL stats in the background.
	click := &db.ShortURLClick{
		IP:         c.IP(),
		Browser:    ua.Name,
		Device:     ua.Device,
		DeviceType: ua.DeviceType(),
		Timestamp:  time.Now().Unix(),
	}

	// Update cache
	s.urlMtx.Lock()
	if _, found = s.urlCache[shortUrl]; found {
		s.urlCache[shortUrl].Clicks++
	}
	s.urlMtx.Unlock()

	defer func() {
		err := s.db.UpdateShortURL(shortUrl, "", click)
		if err != nil {
			appLog.Printf("\ndb.UpdateShortURL error: %v\n", err)
		}
	}()

	return c.Redirect(urlInfo.OriginalURL, codeFound)
}

// handleURLUpdate handles the "PATCH /api/url?shortUrl="short-url" endpoint and
// updates the short URL in the query.
func (s *WebServer) handleURLUpdate(c *fiber.Ctx) error {
	if _, ok := c.Context().UserValue(ctxID).(string); !ok {
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortURL := c.Query("shortUrl")
	if shortURL == "" {
		return errBadRequest("invalid short URL")
	}

	var form *updateShortURLRequest
	if err := c.BodyParser(&form); err != nil {
		return errBadRequest("invalid request body")
	}

	if form.Disable == nil && form.LongURL == "" {
		return errBadRequest("missing required fields")
	}

	if form.LongURL != "" {
		longURL, err := url.ParseRequestURI(form.LongURL)
		if err != nil || longURL.Scheme != "https" || longURL.Host == "" {
			return errBadRequest("invalid URL, provide an absolute URL with a scheme (only https is allowed) and a host (e.g. https://example.com/path/to/resource))")
		}

		if err := s.db.UpdateShortURL(shortURL, form.LongURL, nil); err != nil {
			return translateDBError(err)
		}

		// Update cache
		s.urlMtx.Lock()
		if _, found := s.urlCache[shortURL]; found {
			s.urlCache[shortURL].OriginalURL = form.LongURL
		}
		s.urlMtx.Unlock()

	} else {
		disable := *form.Disable
		if err := s.db.ToggleShortLinkStatus(shortURL, disable); err != nil {
			translateDBError(err)
		}

		// Update cache
		s.urlMtx.Lock()
		if _, found := s.urlCache[shortURL]; found {
			s.urlCache[shortURL].Disabled = disable
		}
		s.urlMtx.Unlock()
	}

	return c.Status(codeOk).JSON(newAPIResponse(true, codeOk, "Short URL has been updated"))
}

// handleGetShortURLClicks handles the "GET /api/url/clicks?shortUrl="short-url"
// endpoint and return the full information for a short url clicks.
func (s *WebServer) handleGetShortURLClicks(c *fiber.Ctx) error {
	if _, ok := c.Context().UserValue(ctxID).(string); !ok {
		return errUnauthorized("you are not unauthorized to access this resource")
	}

	shortURL := c.Query("shortUrl")
	if shortURL == "" {
		return errBadRequest("invalid short URL")
	}

	clicks, err := s.db.RetrieveShortURLClicks(shortURL)
	if err != nil {
		return translateDBError(err)
	}

	resp := &struct {
		*APIResponse
		Data []*db.ShortURLClick `json:"data"`
	}{
		APIResponse: newAPIResponse(true, codeOk, "Short URL clicks retrieved"),
		Data:        clicks,
	}

	return c.Status(codeOk).JSON(resp)
}
