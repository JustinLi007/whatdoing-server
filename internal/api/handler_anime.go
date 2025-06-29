package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerAnime interface {
	NewAnime(w http.ResponseWriter, r *http.Request)
	GetAnime(w http.ResponseWriter, r *http.Request)
	GetAllAnime(w http.ResponseWriter, r *http.Request)
}

type handlerAnime struct {
	dbsAnime         database.DbsAnime
	dbsRelUsersAnime database.DbsRelUsersAnime
}

var handlerAnimeInstance *handlerAnime

func NewHandlerAnime(dbsAnime database.DbsAnime, dbsRelUsersAnime database.DbsRelUsersAnime) HandlerAnime {
	if handlerAnimeInstance != nil {
		return handlerAnimeInstance
	}

	newHandlerAnime := &handlerAnime{
		dbsAnime:         dbsAnime,
		dbsRelUsersAnime: dbsRelUsersAnime,
	}
	handlerAnimeInstance = newHandlerAnime

	return handlerAnimeInstance
}

func (h *handlerAnime) NewAnime(w http.ResponseWriter, r *http.Request) {
	type AnimeRequest struct {
		AnimeId     *string `json:"anime_id"`
		Name        *string `json:"name"`
		ContentType *string `json:"content_type"`
		Description *string `json:"description"`
		ImageUrl    *string `json:"image_url"`
		Episodes    *int    `json:"episodes"`
	}

	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: handler anime NewAnime GetUser: user is nil")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	var req AnimeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: handler anime NewAnime req decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	anime := &database.Anime{
		AnimeName: database.AnimeName{},
	}
	var dbAnime *database.Anime

	if req.ContentType == nil {
		log.Printf("error: handler anime NewAnime: content type missing")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	// TODO: maybe have consts
	if strings.TrimSpace(strings.ToLower(*req.ContentType)) != "anime" {
		log.Printf("error: handler anime NewAnime: invalid content type")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	if req.AnimeId != nil {
		err := uuid.Validate(*req.AnimeId)
		if err != nil {
			log.Printf("error: handler anime NewAnime: validate anime id")
			utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
				"error": "internal server error",
			})
			return
		}
		animeId, err := uuid.Parse(*req.AnimeId)
		if err != nil {
			log.Printf("error: handler anime NewAnime: validate anime id")
			utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
				"error": "internal server error",
			})
			return
		}
		anime.Id = animeId
		dbAnime, err = h.dbsAnime.GetAnimeById(anime)
		if err != nil {
			log.Printf("error: handler anime NewAnime: GetAnimeById: %v", err)
			utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
				"error": "internal server error",
			})
			return
		}
	} else {
		if req.Name == nil {
			log.Printf("error: handler anime NewAnime: missing name")
			utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
				"error": "internal server error",
			})
			return
		}

		anime.Episodes = req.Episodes
		anime.Description = req.Description
		anime.AnimeName.Name = *req.Name
		dbAnime, err = h.dbsAnime.InsertAnime(anime)
		if err != nil {
			log.Printf("error: handler anime NewAnime: InsertAnime: %v", err)
			utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
				"error": "internal server error",
			})
			return
		}
	}

	relUserAnime := &database.RelUsersAnime{
		UserId:  user.Id,
		AnimeId: dbAnime.Id,
	}
	err = h.dbsRelUsersAnime.InsertRel(relUserAnime)
	if err != nil {
		log.Printf("error: handler anime NewAnime: InsertRel: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"next": fmt.Sprintf("/contents/%s", dbAnime.Id),
	})
}

func (h *handlerAnime) GetAnime(w http.ResponseWriter, r *http.Request) {
	type GetAnimeRequest struct {
		ContentId *string `json:"content_id"`
	}

	anime := &database.Anime{}
	dbAnime, err := h.dbsAnime.GetAnimeById(anime)
	if err != nil {
		log.Printf("error: handler anime GetAnimeById: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime": dbAnime,
	})
}

func (h *handlerAnime) GetAllAnime(w http.ResponseWriter, r *http.Request) {
	dbAnimeList, err := h.dbsAnime.GetAllAnime()
	if err != nil {
		log.Printf("error: handler anime GetAllAnime: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime_list": dbAnimeList,
	})
}
