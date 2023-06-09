package api

import "net/http"

// handleCreateLink handles the POST /links endpoint and creates a new short
// URL.
func (r *Router) handleCreateLink(w http.ResponseWriter, req *http.Request) {

}

// handleGetLink handles the GET /links/{shortUrl} endpoint and returns the full
// information about a short URL for a validated user.
func (r *Router) handleGetLink(w http.ResponseWriter, req *http.Request) {
	

}

// handleGetAllLinks handles the GET /links endpoint and returns all the short
// URLs for a validated user.
func (r *Router) handleGetAllLinks(w http.ResponseWriter, req *http.Request) {

}

// handleCreateLinkQR handles the GET /links/{shortUrl}/qr endpoint and returns
// a QR code for the short URL.
func (r *Router) handleCreateLinkQR(w http.ResponseWriter, req *http.Request) {

}
