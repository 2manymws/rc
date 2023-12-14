package rfc9111

import "errors"

var ErrNegativeRatio = errors.New("invalid heuristic expiration ratio (< 0)")
