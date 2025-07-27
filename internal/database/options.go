package database

import (
	"github.com/google/uuid"
)

const (
	STARTED     = "started"
	NOT_STARTED = "not-started"
	COMPLETED   = "completed"
)

type options struct {
	relId       uuid.UUID
	animeId     uuid.UUID
	status      string
	withRelId   bool
	withAnimeId bool
}

func newOptions() *options {
	options := &options{
		status:    "",
		withRelId: false,
	}
	return options
}

type OptionsFunc func(o *options)

func WithRelId(id uuid.UUID) OptionsFunc {
	return func(o *options) {
		o.withRelId = true
		o.relId = id
	}
}

func WithAnimeId(id uuid.UUID) OptionsFunc {
	return func(o *options) {
		o.withAnimeId = true
		o.animeId = id
	}
}

func WithStatus(status string) OptionsFunc {
	return func(o *options) {
		switch status {
		case STARTED:
			o.status = STARTED
		case NOT_STARTED:
			o.status = NOT_STARTED
		case COMPLETED:
			o.status = COMPLETED
		default:
			o.status = ""
		}
	}
}
