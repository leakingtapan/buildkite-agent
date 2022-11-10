package api

import (
	"errors"
	"fmt"
)

var ErrAudienceTooLong = errors.New("the API only supports at most one element in the audience")

type OidcToken struct {
	Token string `json:"token"`
}

func (c *Client) OidcToken(jobId string, audience ...string) (*OidcToken, *Response, error) {
	type oidcTokenRequest struct {
		Audience string `json:"audience"`
	}

	var m *oidcTokenRequest
	switch len(audience) {
	case 0:
		m = nil
	case 1:
		m = &oidcTokenRequest{Audience: audience[0]}
	default:
		return nil, nil, ErrAudienceTooLong
	}

	u := fmt.Sprintf("jobs/%s/oidc/tokens", jobId)
	req, err := c.newRequest("POST", u, m)
	if err != nil {
		return nil, nil, err
	}

	t := &OidcToken{}
	resp, err := c.doRequest(req, t)
	if err != nil {
		return nil, nil, err
	}

	return t, resp, err
}
