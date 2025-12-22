//go:build sms
// +build sms

package sms

import (
	"context"
	"errors"

	"access_device/auth"
)

type Strategy struct {
	Phone string
	Code  string
	Auth  auth.Authenticator
}

func (s *Strategy) Name() string { return "sms" }

func (s *Strategy) Authenticate(ctx context.Context) (bool, *auth.Result, error) {
	if s.Phone == "" {
		return false, nil, errors.New("missing phone")
	}
	if s.Code == "" {
		return false, nil, errors.New("missing sms code")
	}

	// TODO: 这里以后可以换成真正的 SMS 校验
	// 目前示例：直接拒绝或 mock
	return false, nil, errors.New("sms auth not implemented")
}
