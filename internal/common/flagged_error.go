package common

import "errors"

type Flag string

const (
	NotExistsFlag Flag = "not exists"
)

type Flagged interface {
	error
	Flag() Flag
}

func FlagError(err error, flag Flag) Flagged {
	return fault{error: err, flag: flag}
}

type fault struct {
	error
	flag Flag
}

func (e fault) Unwrap() error {
	return e.error
}

func (e fault) Flag() Flag {
	return e.flag
}

func IsFlaggedError(err error, flag Flag) bool {
	if err == nil {
		return false
	}

	var flagged Flagged
	if errors.As(err, &flagged) {
		return flagged.Flag() == flag
	}

	return false
}
