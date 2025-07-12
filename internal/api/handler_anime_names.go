package api

import (
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerAnimeNames interface {
	GetNamesByAnime(w http.ResponseWriter, r *http.Request)
	GetNames(w http.ResponseWriter, r *http.Request)
}

type handlerAnimeNames struct {
	dbsAnimeNames         database.DbsAnimeNames
	dbsRelAnimeAnimeNames database.DbsRelAnimeAnimeNames
}

var handlerAnimeNamesInstance *handlerAnimeNames

func NewHandlerAnimeNames(dbsAnimeNames database.DbsAnimeNames, dbsRelAnimeAnimeNames database.DbsRelAnimeAnimeNames) HandlerAnimeNames {
	if handlerAnimeNamesInstance != nil {
		return handlerAnimeNamesInstance
	}

	newHandlerAnimeNames := &handlerAnimeNames{
		dbsAnimeNames:         dbsAnimeNames,
		dbsRelAnimeAnimeNames: dbsRelAnimeAnimeNames,
	}
	handlerAnimeNamesInstance = newHandlerAnimeNames

	return handlerAnimeNamesInstance
}

func (h *handlerAnimeNames) GetNamesByAnime(w http.ResponseWriter, r *http.Request) {
	animeIdStr := r.PathValue("animeId")

	err := uuid.Validate(animeIdStr)
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNamesByAnime: validate id str: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	animeId, err := uuid.Parse(animeIdStr)
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNamesByAnime: parse id str: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	animeReq := &database.Anime{
		Id: animeId,
	}
	dbAnimeNames, err := h.dbsAnimeNames.GetNamesByAnime(animeReq)
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNamesByAnime: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime_names": dbAnimeNames,
	})
}

func (h *handlerAnimeNames) GetNames(w http.ResponseWriter, r *http.Request) {
	dbRelAnimeAnimeNames, err := h.dbsRelAnimeAnimeNames.GetNames()
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNames: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"raan": dbRelAnimeAnimeNames,
	})
}
