package model

import "errors"

var (
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrUserNotFound        = errors.New("user not found")
	ErrLinkNotFound        = errors.New("link not found")
	ErrLinkAlreadyExists   = errors.New("link already exists")
	ErrInvalidURL          = errors.New("invalid url")
	ErrInternalError       = errors.New("internal error")
	ErrEncodingOriginalURL = errors.New("encoding original url failed")
	ErrDecodingShortURL    = errors.New("decoding short url failed")
)
