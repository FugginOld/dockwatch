package types

// TokenResponse is returned by the registry on successful authentication
type TokenResponse struct {
	Token string `json:"token"`
	// AccessToken is what the OAuth2 form of the token endpoint returns. Azure ACR
	// and some GitLab configurations send only this field, and reading just "token"
	// left the bearer empty -- which surfaced as every digest check failing over to
	// a full pull, silently, on every poll.
	AccessToken string `json:"access_token"`
}

// Bearer returns the token the registry issued, whichever field it used.
func (r TokenResponse) Bearer() string {
	if r.Token != "" {
		return r.Token
	}
	return r.AccessToken
}
