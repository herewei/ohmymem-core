package domain

import "errors"

var (
	ErrInvalidCategory  = errors.New("invalid category")
	ErrInvalidTag       = errors.New("invalid tag")
	ErrInvalidContent   = errors.New("invalid content")
	ErrInvalidRationale = errors.New("invalid rationale")
	ErrForbiddenContent = errors.New("forbidden content")
	ErrListItem         = errors.New("list item not allowed")
)
