package webserver

import (
	"fmt"
	"time"

	"github.com/cristalhq/jwt/v4"
)

const (
	// jwtIssuer is the JWT issuer.
	jwtIssuer = "B.O.B"
	// jwtAlg is the JWT algorithm to use.
	jwtAlg = jwt.HS256
	// jwtAudienceUser is the JWT audience for user authentication.
	jwtAudienceUser = "jwt-user"
	// tokenExpiry is the default JWT token expiry.
	tokenExpiry = 24 * time.Hour
)

// jwtAudience is the JWT audience type.
type jwtAudience string

// jwtAuthenticator is the JWT authenticator. It is used to generate and
// validate JWT tokens.
type jwtAuthenticator struct {
	builder  *jwt.Builder
	verifier jwt.Verifier
}

// newJWTAuthenticator creates a new *jwtAuthenticator.
func newJWTAuthenticator(alg jwt.Algorithm) (*jwtAuthenticator, error) {
	jwtSecret, err := randomBytes(32)
	if err != nil {
		return nil, fmt.Errorf("RandomBytes error: %w", err)
	}

	signer, err := jwt.NewSignerHS(alg, jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("jwt.NewSignerHS error: %w", err)
	}

	verifier, err := jwt.NewVerifierHS(alg, jwtSecret[:])
	if err != nil {
		return nil, fmt.Errorf("jwt.NewVerifierHS error: %w", err)
	}

	authenticator := &jwtAuthenticator{
		builder:  jwt.NewBuilder(signer),
		verifier: verifier,
	}

	return authenticator, nil
}

// generateAuthToken generates a new JWT token.
func (jwtAuth *jwtAuthenticator) generateAuthToken(id, subject string, audience jwtAudience, expiry time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ID:        id,
		Audience:  []string{string(audience)},
		Subject:   subject,
		Issuer:    jwtIssuer,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
	}

	token, err := jwtAuth.builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("jwtAuthenticator.builder.Build error: %w", err)
	}

	return token.String(), nil
}

// validateAuthToken validates the given JWT token.
func (jwtAuth *jwtAuthenticator) validateAuthToken(token string) (*jwt.RegisteredClaims, bool) {
	jwtClaims := new(jwt.RegisteredClaims)
	err := jwt.ParseClaims([]byte(token), jwtAuth.verifier, jwtClaims)
	if err != nil || !isValidJWTClaims(jwtClaims) {
		return nil, false
	}

	return jwtClaims, true
}

// isValidJWTClaims checks if the given JWT claims are valid.
func isValidJWTClaims(jwtClaims *jwt.RegisteredClaims) bool {
	return jwtClaims.IsIssuer(jwtIssuer) && jwtClaims.IsValidAt(time.Now()) && jwtClaims.IsForAudience(jwtAudienceUser)
}
