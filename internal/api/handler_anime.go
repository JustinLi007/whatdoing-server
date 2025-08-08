package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerAnime interface {
	NewAnime(w http.ResponseWriter, r *http.Request)
	GetAnime(w http.ResponseWriter, r *http.Request)
	GetAllAnime(w http.ResponseWriter, r *http.Request)
	UpdateAnime(w http.ResponseWriter, r *http.Request)
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
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: Anime: NewAnime: GetUser: user is nil")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: GetUser: WriteJson: %v", err)
		}
		return
	}

	type AnimeRequest struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		ImageUrl    *string `json:"image_url"`
		Episodes    *int    `json:"episodes"`
	}

	var req AnimeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: Handler: Anime: NewAnime: Decode: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: Decode: WriteJson: %v", err)
		}
		return
	}

	if req.Name == nil {
		log.Printf("error: Handler: Anime: NewAnime: missing name: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: missing name: WriteJson: %v", err)
		}
		return
	}
	if req.Episodes == nil {
		log.Printf("error: Handler: Anime: NewAnime: missing episodes: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: missing episodes: WriteJson: %v", err)
		}
		return
	}
	if req.ImageUrl == nil {
		log.Printf("error: Handler: Anime: NewAnime: missing image url: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: missing image url: WriteJson: %v", err)
		}
		return
	}
	if req.Description == nil {
		log.Printf("error: Handler: Anime: NewAnime: missing description: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: missing description: WriteJson: %v", err)
		}
		return
	}

	reqAnime := &database.Anime{
		Episodes:    req.Episodes,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		AnimeName: database.AnimeName{
			Name: *req.Name,
		},
	}

	dbAnime, err := h.dbsAnime.InsertAnime(reqAnime)
	if err != nil {
		log.Printf("error: Handler: Anime: NewAnime: InsertAnime: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: NewAnime: InsertAnime: WriteJson: %v", err)
		}
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime": dbAnime,
	})
}

func (h *handlerAnime) GetAnime(w http.ResponseWriter, r *http.Request) {
	contentIdStr := r.PathValue("contentId")

	if contentIdStr == "" {
		log.Printf("error: handler anime GetAnime: missing content id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	err := uuid.Validate(contentIdStr)
	if err != nil {
		log.Printf("error: handler anime GetAnime: validate uuid: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	id, err := uuid.Parse(contentIdStr)
	if err != nil {
		log.Printf("error: handler anime GetAnime: parse uuid: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	anime := &database.Anime{
		Id: id,
	}
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
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: Anime: GetAllAnime: GetUser: user nil")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: GetAllAnime: GetUser: WriteJson: %v", err)
		}
		return
	}

	opts := make([]database.OptionsFunc, 0)
	queries := r.URL.Query()

	search := queries.Get("search")
	sort := queries.Get("sort")
	ignore := queries.Get("ignore")
	opts = append(opts, database.WithSearch(search))
	opts = append(opts, database.WithSort(sort))
	opts = append(opts, database.WithIgnore(ignore))

	log.Printf("search: '%v'", search)
	log.Printf("sort: '%v'", sort)
	log.Printf("ignore: '%v'", ignore)

	dbAnimeList, err := h.dbsAnime.GetAllAnime(user, opts...)
	if err != nil {
		log.Printf("error: Handler: Anime: GetAllAnime: GetAllAnime: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: Anime: GetAllAnime: GetAllAnime: WriteJson: %v", err)
		}
		return
	}

	if err = utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"anime": dbAnimeList,
	}); err != nil {
		log.Printf("error: Handler: Anime: GetAllAnime: payload: WriteJson: %v", err)
	}
}

func (h *handlerAnime) UpdateAnime(w http.ResponseWriter, r *http.Request) {
	type AnimeRequest struct {
		ContentId        *string  `json:"content_id"`
		ContentNamesId   *string  `json:"content_names_id"`
		ContentType      *string  `json:"content_type"`
		Description      *string  `json:"description"`
		ImageUrl         *string  `json:"image_url"`
		Episodes         *int     `json:"episodes"`
		AlternativeNames []string `json:"alternative_names"`
	}

	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: handler anime UpdateAnime GetUser: user is nil")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	var req AnimeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime req decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	if req.ContentType == nil {
		log.Printf("error: handler anime UpdateAnime: content type missing")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	if strings.TrimSpace(strings.ToLower(*req.ContentType)) != "anime" {
		log.Printf("error: handler anime UpdateAnime: invalid content type")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	if req.ContentId == nil {
		log.Printf("error: handler anime UpdateAnime: missing content id")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	if req.ContentNamesId == nil {
		log.Printf("error: handler anime UpdateAnime: missing content name id")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "status bad request",
		})
		return
	}

	err = uuid.Validate(*req.ContentId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: validate content id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	animeId, err := uuid.Parse(*req.ContentId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: parse content id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	err = uuid.Validate(*req.ContentNamesId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: validate content name id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	animeNamesId, err := uuid.Parse(*req.ContentNamesId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: parse content name id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	anime := &database.Anime{
		Id:          animeId,
		UpdatedAt:   time.Now(),
		Episodes:    req.Episodes,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		AnimeName: database.AnimeName{
			Id: animeNamesId,
		},
		AlternativeNames: make([]*database.AnimeName, 0),
	}

	for _, v := range req.AlternativeNames {
		if strings.TrimSpace(v) == "" {
			continue
		}
		anime.AlternativeNames = append(anime.AlternativeNames, &database.AnimeName{
			Name: v,
		})
	}

	err = h.dbsAnime.UpdateAnime(anime)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{})
}
