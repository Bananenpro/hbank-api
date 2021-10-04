package services

import "errors"

var (
	ErrEmailExists           = errors.New("email-exists")
	ErrInvalidCaptchaToken   = errors.New("invalid-captcha-token")
	ErrNotFound              = errors.New("not-found")
	ErrEmailAlreadyConfirmed = errors.New("email-already-confirmed")
	ErrInvalidCredentials    = errors.New("invalid-credentials")
	ErrTimeout               = errors.New("timeout")
)
