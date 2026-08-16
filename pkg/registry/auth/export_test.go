package auth

import (
	"context"
	"net/http"

	ref "github.com/distribution/reference"
)

// BearerFromResponseForTest exposes bearerFromResponse to the external test package.
func BearerFromResponseForTest(res *http.Response) (string, error) { return bearerFromResponse(res) }

// TokenForChallengeForTest exposes tokenForChallenge to the external test package.
func TokenForChallengeForTest(ctx context.Context, header string, status int, imageRef ref.Named, registryAuth string) (string, error) {
	return tokenForChallenge(ctx, header, status, imageRef, registryAuth)
}
