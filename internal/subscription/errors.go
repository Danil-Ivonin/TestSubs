package subscription

import "errors"

var (
	ErrInvalidDate   = errors.New("invalid date")
	ErrInvalidPeriod = errors.New("invalid period")
	ErrInvalidPrice  = errors.New("invalid price")
	ErrInvalidUUID   = errors.New("invalid uuid")
	ErrNotFound      = errors.New("subscription not found")
)
