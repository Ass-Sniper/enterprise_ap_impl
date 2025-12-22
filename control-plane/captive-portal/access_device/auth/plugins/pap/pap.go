//go:build pap
// +build pap

package pap

import (
	"context"
	"errors"

	"access_device/auth"
)

type Strategy struct {
	Username string
	Password string
	Auth     auth.Authenticator
}

func (p *Strategy) Name() string { return "pap" }

func (p *Strategy) Authenticate(ctx context.Context) (bool, *auth.Result, error) {
	if p.Username == "" {
		return false, nil, errors.New("missing username")
	}
	if p.Password == "" {
		return false, nil, errors.New("missing password")
	}
	if p.Auth == nil {
		return false, nil, errors.New("radius authenticator not set")
	}

	ok, reply, timeout, err :=
		p.Auth.RadiusPAP(ctx, p.Username, p.Password)
	if err != nil {
		return false, nil, err
	}

	res := &auth.Result{
		Username:       p.Username,
		AuthMethod:     "pap",
		SessionTimeout: timeout,
		ReplyAttrs:     reply,
	}

	if v := reply["Filter-Id"]; v != "" {
		res.Policy = v
	}

	return ok, res, nil
}
