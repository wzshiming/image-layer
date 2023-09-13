package registry

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/distribution/distribution/v3/registry/client/auth/challenge"
)

type userPass struct {
	Username string
	Password string
}

type basicCredentials struct {
	credentials map[string]userPass
}

func newBasicCredentials(up userPass, domain string, insecure bool) (*basicCredentials, error) {
	bc := &basicCredentials{
		credentials: map[string]userPass{},
	}

	scheme := "https"
	if insecure {
		scheme = "http"
	}
	urls, err := getAuthURLs(scheme + "://" + domain)
	if err != nil {
		return nil, err
	}
	for _, u := range urls {
		bc.credentials[u] = up
	}

	return bc, nil
}

func (c *basicCredentials) Basic(u *url.URL) (string, string) {
	up := c.credentials[u.String()]

	return up.Username, up.Password
}

func (c *basicCredentials) RefreshToken(u *url.URL, service string) string {
	return ""
}

func (c *basicCredentials) SetRefreshToken(u *url.URL, service, token string) {
}

func getAuthURLs(remoteURL string) ([]string, error) {
	authURLs := []string{}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return nil, err
	}

	remoteURL = u.String()

	resp, err := http.Get(remoteURL + "/v2/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	for _, c := range challenge.ResponseChallenges(resp) {
		if strings.EqualFold(c.Scheme, "bearer") {
			authURLs = append(authURLs, c.Parameters["realm"])
		}
	}

	return authURLs, nil
}
