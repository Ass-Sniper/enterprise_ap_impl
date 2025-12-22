//go:build token
// +build token

package token

import (
	"context"
	"errors"

	"access_device/auth"
)

type Strategy struct {
	Username string
	Token    string
}

func (t *Strategy) Name() string { return "token" }

func (t *Strategy) Authenticate(ctx context.Context) (bool, *auth.Result, error) {
	if t.Username == "" {
		return false, nil, errors.New("missing username")
	}
	if t.Token == "" {
		return false, nil, errors.New("missing token")
	}

	// TODO: 换成真实 Token 校验（JWT / Redis / HTTP）
	if t.Token != "VALID_TOKEN" {
		return false, &auth.Result{
			Username:   t.Username,
			AuthMethod: "token",
		}, nil
	}

	return true, &auth.Result{
		Username:   t.Username,
		AuthMethod: "token",
		Policy:     "staff",
	}, nil
}
