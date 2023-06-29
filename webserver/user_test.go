package webserver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

var dummyUserPassword = "password"

func TestWebServer_handleUsernameExists(t *testing.T) {
	s := newTServer(t)
	defer s.Stop()

	tests := []struct {
		name, username string
		wantExists     bool
		dbErr          error
	}{{
		name:       "valid username exists",
		username:   "fibrealz",
		wantExists: true,
	}, {
		name:     "invalid username",
		username: "fi",
	}, {
		name:     "valid username does not exist",
		username: "fibrealz",
	}, {
		name:     "server error",
		username: "fibrealz",
		dbErr:    dummyError,
	}}

	for _, tt := range tests {
		tdb := s.db.(*tDB)
		tdb.usernameExists = tt.wantExists
		tdb.err = tt.dbErr

		var resp *usernameExitsResponse
		err := s.sendRequest(fiber.MethodGet, fmt.Sprintf("api/username-exists?username=%s", tt.username), nil, &resp, nil)
		if err != nil {
			t.Fatalf("%s: s.sendRequest error: %s", tt.name, err)
		}

		if resp == nil || resp.APIResponse == nil {
			t.Fatalf("%s: Expected an API response but got nothing", tt.name)
		}

		if tt.wantExists != resp.Exists {
			t.Fatalf("%s: Expected %t got %t", tt.name, tt.wantExists, resp.Exists)
		}

		if tt.dbErr != nil && resp.Code != codeInternal {
			t.Fatalf("%s: Expected server error got %v", tt.name, resp.Code)
		}
	}
}

func TestWebServer_handleCreateAccount(t *testing.T) {
	s := newTServer(t)
	defer s.Stop()

	tests := []struct {
		name          string
		req           createAccountRequest
		messagePrefix string
		dbErr         error
	}{{
		name: "success",
		req: createAccountRequest{
			Username: "fibrealz",
			Email:    "testmail@example.com",
			Password: dummyUserPassword,
		},
		messagePrefix: "Account Created",
	}, {
		name: "invalid username",
		req: createAccountRequest{
			Username: "fi",
			Email:    "testmail@example.com",
			Password: dummyUserPassword,
		},
		messagePrefix: "username with at least",
	}, {
		name: "invalid email",
		req: createAccountRequest{
			Username: "fibrealz",
			Email:    "testmail@example",
			Password: dummyUserPassword,
		},
		messagePrefix: "a valid email is required",
	}, {
		name: "invalid password length",
		req: createAccountRequest{
			Username: "fibrealz",
			Email:    "testmail@example.com",
			Password: "asdf",
		},
		messagePrefix: "password must be a minimum of",
	}, {
		name: "server error",
		req: createAccountRequest{
			Username: "fibrealz",
			Email:    "testmail@example.com",
			Password: dummyUserPassword,
		},
		messagePrefix: "Something unexpected happened",
		dbErr:         dummyError,
	}}

	for _, tt := range tests {
		tdb := s.db.(*tDB)
		tdb.err = tt.dbErr

		var resp *APIResponse
		err := s.sendRequest(fiber.MethodPost, "api/user", tt.req, &resp, nil)
		if err != nil {
			t.Fatalf("%s: s.sendRequest error: %s", tt.name, err)
		}

		if resp == nil {
			t.Fatalf("%s: Expected an API response but got nothing", tt.name)
		}

		if !strings.Contains(resp.Message, tt.messagePrefix) {
			t.Fatalf("%s: Expected a response message that to contains %s but got %s", tt.name, tt.messagePrefix, resp.Message)
		}

		if tt.dbErr != nil && resp.Code != codeInternal {
			t.Fatalf("%s: Expected server error got %v", tt.name, resp)
		}

		// Confirm DB changes.
		if resp.Ok {
			if exists, err := s.db.UsernameExists(tt.req.Username); err != nil {
				t.Fatalf("%s: Expected db.UsernameExists error: %v", tt.name, resp)
			} else if !exists {
				t.Fatalf("%s: Newly created user does not exist", tt.name)
			}
		}
	}
}

func TestWebServer_handleGetUser(t *testing.T) {
	s := newTServer(t)
	defer s.Stop()

	userEmail := "test@email.com"
	userUsername := "fibrealz"
	authToken, err := s.authenticator.generateAuthToken(userEmail, userUsername, jwtAudienceUser, tokenExpiry)
	if err != nil {
		t.Fatalf("s.authenticator.generateAuthToken error: %s", err)
	}

	tests := []struct {
		name             string
		apiKey           string
		wantUnAuthorized bool
		dbErr            error
	}{{
		name:   "success",
		apiKey: authToken,
	}, {
		name:             "unauthorized: missing api key",
		wantUnAuthorized: true,
	}, {
		name:             "unauthorized: invalid api key",
		apiKey:           authToken[:len(authToken)-5],
		wantUnAuthorized: true,
	}, {
		name:   "server error",
		apiKey: authToken,
		dbErr:  dummyError,
	}}

	for _, tt := range tests {
		tdb := s.db.(*tDB)
		tdb.err = tt.dbErr

		headers := map[string]string{
			fiber.HeaderAuthorization: fmt.Sprintf("Bearer %s", tt.apiKey),
		}

		var resp *userInfoResponse
		if err := s.sendRequest(fiber.MethodGet, "api/user", nil, &resp, headers); err != nil {
			t.Fatalf("%s: s.sendRequest error: %s", tt.name, err)
		}

		if resp == nil {
			t.Fatalf("%s: Expected a response but got nothing", tt.name)
		}

		if tt.wantUnAuthorized && resp.Code != codeUnauthorized {
			t.Fatalf("%s: Expected unauthorized error got %v", tt.name, resp.Code)
		}

		if tt.dbErr == nil && !tt.wantUnAuthorized && resp.Code != codeOk {
			t.Fatalf("%s: Expected OK got %v", tt.name, resp.Code)
		}

		if tt.dbErr != nil && resp.Code != codeInternal {
			t.Fatalf("%s: Expected server error got %v", tt.name, resp.Code)
		}

		if (resp.Ok && resp.Data == nil) || (resp.Ok && tdb.dummyUser.Email != resp.Data.Email) {
			t.Fatalf("%s: Expected matching user info", tt.name)
		}
	}
}

func TestWebServer_handleLogin(t *testing.T) {
	s := newTServer(t)
	defer s.Stop()

	tests := []struct {
		name          string
		req           loginRequest
		messagePrefix string
		dbErr         error
	}{{
		name: "success",
		req: loginRequest{
			Email:    "test@email.com",
			Password: dummyUserPassword,
		},
		messagePrefix: "Login Successful",
	}, {
		name: "invalid email",
		req: loginRequest{
			Email:    "testmail@example",
			Password: dummyUserPassword,
		},
		messagePrefix: "a valid email is required",
	}, {
		name: "invalid password length",
		req: loginRequest{
			Email: "test@email.com",
		},
		messagePrefix: "password is required",
	}, {
		name: "incorrect password",
		req: loginRequest{
			Email:    "test@email.com",
			Password: "incorrect password",
		},
		messagePrefix: "incorrect password",
	}, {
		name: "server error",
		req: loginRequest{
			Email:    "test@email.com",
			Password: dummyUserPassword,
		},
		messagePrefix: "Something unexpected happened",
		dbErr:         dummyError,
	}}
	for _, tt := range tests {
		tdb := s.db.(*tDB)
		tdb.err = tt.dbErr

		var resp *loginResponse
		if err := s.sendRequest(fiber.MethodPost, "api/login", tt.req, &resp, nil); err != nil {
			t.Fatalf("%s: s.sendRequest error: %s", tt.name, err)
		}

		if resp == nil {
			t.Fatalf("%s: Expected a response but got nothing", tt.name)
		}

		if tt.dbErr != nil && resp.Code != codeInternal {
			t.Fatalf("%s: Expected server error got %v", tt.name, resp.Code)
		}

		if (resp.Ok && resp.Data == nil) || (resp.Ok && tdb.dummyUser.Email != resp.Data.Email) {
			t.Fatalf("%s: Expected matching user info", tt.name)
		}

		if !strings.Contains(resp.Message, tt.messagePrefix) {
			t.Fatalf("%s: Expected a response message that contains %s but got %s", tt.name, tt.messagePrefix, resp.Message)
		}
	}
}
