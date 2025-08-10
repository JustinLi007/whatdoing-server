package database

import (
	"strings"

	"github.com/google/uuid"
)

const (
	STATUS_STARTED     = "started"
	STATUS_NOT_STARTED = "not-started"
	STATUS_COMPLETED   = "completed"

	SORT_ASC  = "ASC"
	SORT_DESC = "DESC"
)

type ProgressId struct {
	Id uuid.UUID
}

type AnimeId struct {
	Id uuid.UUID
}

type ProgressStatus struct {
	StatusValue string
}

type Search struct {
	SearchValue string
}

type Sort struct {
	SortValue string
}

type Options struct {
	ProgressId      *ProgressId
	AnimeId         *AnimeId
	Status          *ProgressStatus
	Search          *Search
	Sort            *Sort
	IgnoreInLibrary bool
}

func NewOptions() *Options {
	options := &Options{
		ProgressId:      nil,
		AnimeId:         nil,
		Status:          nil,
		Search:          nil,
		Sort:            nil,
		IgnoreInLibrary: false,
	}
	return options
}

type OptionsFunc func(o *Options)

func WithProgressId(id uuid.UUID) OptionsFunc {
	return func(o *Options) {
		o.ProgressId = &ProgressId{
			Id: id,
		}
	}
}

func WithAnimeId(id uuid.UUID) OptionsFunc {
	return func(o *Options) {
		o.AnimeId = &AnimeId{
			Id: id,
		}
	}
}

func WithStatus(value string) OptionsFunc {
	status_value := strings.ToLower(strings.TrimSpace(value))
	return func(o *Options) {
		switch status_value {
		case STATUS_STARTED:
			o.Status = &ProgressStatus{
				StatusValue: STATUS_STARTED,
			}
		case STATUS_NOT_STARTED:
			o.Status = &ProgressStatus{
				StatusValue: STATUS_NOT_STARTED,
			}
		case STATUS_COMPLETED:
			o.Status = &ProgressStatus{
				StatusValue: STATUS_COMPLETED,
			}
		default:
			// nothing
		}
	}
}

func WithIgnore(value string) OptionsFunc {
	ignore_value := strings.ToLower(strings.TrimSpace(value))
	return func(o *Options) {
		switch ignore_value {
		case "library":
			o.IgnoreInLibrary = true
		default:
			// nothing
		}
	}
}

func WithSearch(value string) OptionsFunc {
	search_value := strings.ToLower(strings.TrimSpace(value))
	return func(o *Options) {
		if search_value == "" {
			return
		}

		o.Search = &Search{
			SearchValue: value,
		}
	}
}

func WithSort(value string) OptionsFunc {
	sort_value := strings.ToUpper(strings.TrimSpace(value))
	return func(o *Options) {
		switch sort_value {
		case SORT_ASC:
			o.Sort = &Sort{
				SortValue: SORT_ASC,
			}
		case SORT_DESC:
			o.Sort = &Sort{
				SortValue: SORT_DESC,
			}
		default:
			// nothing
		}
	}
}
