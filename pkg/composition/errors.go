package composition

import "errors"

var (
	ErrEmptyInput      = errors.New("either image or product_name must be provided")
	ErrConflictingData = errors.New("provide either image or product_name, not both")
	ErrNoComposition   = errors.New("no composition items found")
)
