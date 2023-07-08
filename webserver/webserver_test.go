package webserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/ukane-philemon/bob/db/mem"
)

var (
	tCtx context.Context

	dummyError = errors.New("dummy error")
)

type tServer struct {
	*WebServer
}

// newTServer creates and starts a new server instance. Callers should
// *WebServer.Stop to shutdown the server.
func newTServer(t *testing.T) *tServer {
	var port int
	for port == 0 {
		port = rand.Intn(65535)
	}

	cfg := Config{
		Host: "127.0.0.1",
		Port: fmt.Sprintf("%d", port),
	}

	// Create a new server.
	s, err := New(tCtx, cfg, mem.New())
	if err != nil {
		t.Fatalf("Error creating server: %v", err)
	}

	// Start the server.
	go s.Start()

	return &tServer{s}
}

// sendRequest mimics an actual http request to the server and unmarshals the
// request result into resp. resp must be a pointer to a struct type.
func (ts *tServer) sendRequest(method string, endpoint string, reqBody interface{}, resp interface{}, headers map[string]string) error {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	req := a.Request()
	req.SetRequestURI(fmt.Sprintf("http://%s/%s", ts.addr, endpoint))
	req.Header.SetMethod(method)
	if reqBody != nil {
		req.Header.SetContentType("application/json")
		body, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("json.Marshal error: %w", err)
		}

		req.SetBody(body)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	if err := a.Parse(); err != nil {
		return err
	}

	_, bodyBytes, errs := a.Bytes()
	var errStr string
	for _, err := range errs {
		errStr += err.Error() + ":"
	}

	if errStr != "" {
		return errors.New(strings.TrimSuffix(errStr, ":"))
	}

	return json.Unmarshal(bodyBytes, resp)
}

func TestMain(m *testing.M) {
	var shutdown context.CancelFunc
	tCtx, shutdown = context.WithCancel(context.Background())
	defer shutdown()
	os.Exit(m.Run())
}
