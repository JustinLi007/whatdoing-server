package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerAnime interface {
	NewAnime(w http.ResponseWriter, r *http.Request)
	GetAnime(w http.ResponseWriter, r *http.Request)
	GetAllAnime(w http.ResponseWriter, r *http.Request)
	UpdateAnime(w http.ResponseWriter, r *http.Request)
	DeleteAnime(w http.ResponseWriter, r *http.Request)
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
		AnimeId      *string `json:"anime_id"`
		AnimeNamesId *string `json:"anime_names_id"`
		Description  *string `json:"description"`
		ImageUrl     *string `json:"image_url"`
		Episodes     *int    `json:"episodes"`
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

	if req.AnimeId == nil {
		log.Printf("error: handler anime UpdateAnime: missing anime id")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	if req.AnimeNamesId == nil {
		log.Printf("error: handler anime UpdateAnime: missing anime name id")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	err = uuid.Validate(*req.AnimeId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: validate anime id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	animeId, err := uuid.Parse(*req.AnimeId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: parse anime id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	err = uuid.Validate(*req.AnimeNamesId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: validate anime name id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	animeNamesId, err := uuid.Parse(*req.AnimeNamesId)
	if err != nil {
		log.Printf("error: handler anime UpdateAnime: parse anime name id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	anime := &database.Anime{
		Id:          animeId,
		Episodes:    req.Episodes,
		Description: req.Description,
		ImageUrl:    req.ImageUrl,
		AnimeName: database.AnimeName{
			Id: animeNamesId,
		},
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

func (h *handlerAnime) DeleteAnime(w http.ResponseWriter, r *http.Request) {
	type DeleteAnimeRequest struct {
		AnimeId *string `json:"anime_id"`
	}

	var req DeleteAnimeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error: handler anime DeleteAnime Decode: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: handler anime DeleteAnime Decode: WriteJson: %v", err)
		}
		return
	}

	if req.AnimeId == nil {
		log.Printf("error: handler anime DeleteAnime: missing anime id")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: handler anime DeleteAnime: missing anime id: WriteJson: %v", err)
		}
		return
	}

	if err := uuid.Validate(*req.AnimeId); err != nil {
		log.Printf("error: handler anime DeleteAnime: Validate: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: handler anime DeleteAnime: Validate: WriteJson: %v", err)
		}
		return
	}

	id, err := uuid.Parse(*req.AnimeId)
	if err != nil {
		log.Printf("error: handler anime DeleteAnime: Parse: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: handler anime DeleteAnime: Parse: WriteJson: %v", err)
		}
		return
	}

	reqAnime := &database.Anime{
		Id: id,
	}
	if err := h.dbsAnime.DeleteAnime(reqAnime); err != nil {
		log.Printf("error: handler anime DeleteAnime: DeleteAnime: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: handler anime DeleteAnime: DeleteAnime: WriteJson: %v", err)
		}
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, utils.Envelope{}); err != nil {
		log.Printf("error: handler anime DeleteAnime: Payload: WriteJson: %v", err)
	}
}
