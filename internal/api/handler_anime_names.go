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

func (h *handlerAnimeNames) GetNamesByAnime(w http.ResponseWriter, r *http.Request) {
	animeIdStr := r.PathValue("animeId")

	err := uuid.Validate(animeIdStr)
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNames: validate id str: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	animeId, err := uuid.Parse(animeIdStr)
	if err != nil {
		log.Panicf("error: handlerAnimeNames GetNames: parse id str: %v", err)
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
		log.Panicf("error: handlerAnimeNames GetNames: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime_names": dbAnimeNames,
	})
}
