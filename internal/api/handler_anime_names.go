package api

import (
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
)

type HandlerAnimeNames interface {
	GetNames(w http.ResponseWriter, r *http.Request)
}

type handlerAnimeNames struct {
	dbsAnimeNames database.DbsAnimeNames
}

var handlerAnimeNamesInstance *handlerAnimeNames

func NewHandlerAnimeNames(dbsAnimeNames database.DbsAnimeNames) HandlerAnimeNames {
	if handlerAnimeNamesInstance != nil {
		return handlerAnimeNamesInstance
	}

	newHandlerAnimeNames := &handlerAnimeNames{
		dbsAnimeNames: dbsAnimeNames,
	}
	handlerAnimeNamesInstance = newHandlerAnimeNames

	return handlerAnimeNamesInstance
}

func (h *handlerAnimeNames) GetNames(w http.ResponseWriter, r *http.Request) {
	utils.WriteJson(w, http.StatusNotImplemented, utils.Envelope{
		"error": "not implemented yet",
	})
}
