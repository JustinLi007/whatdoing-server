package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerUserLibraryAnime interface {
	AddToLibrary(w http.ResponseWriter, r *http.Request)
	GetProgress(w http.ResponseWriter, r *http.Request)
	SetProgress(w http.ResponseWriter, r *http.Request)
}

type handlerUserLibraryAnime struct {
	dbsUserLibrary         database.DbsUserLibrary
	dbsRelAnimeUserLibrary database.DbsRelAnimeUserLibrary
}

var handlerUserLibraryAnimeInstance *handlerUserLibraryAnime

func NewHandlerUserLibraryAnime(dbsUserLibrary database.DbsUserLibrary, dbsRelAnimeUserLibrary database.DbsRelAnimeUserLibrary) HandlerUserLibraryAnime {
	if handlerUserLibraryAnimeInstance != nil {
		return handlerUserLibraryAnimeInstance
	}
	newHandlerUserLibraryAnime := &handlerUserLibraryAnime{
		dbsUserLibrary:         dbsUserLibrary,
		dbsRelAnimeUserLibrary: dbsRelAnimeUserLibrary,
	}
	handlerUserLibraryAnimeInstance = newHandlerUserLibraryAnime

	return handlerUserLibraryAnimeInstance
}

func (h *handlerUserLibraryAnime) AddToLibrary(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: GetUser: nil")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	type AddToLibraryRequest struct {
		AnimeId *uuid.UUID `json:"anime_id"`
	}

	var req AddToLibraryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: Decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	if req.AnimeId == nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: missing anime id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	reqAnime := &database.Anime{
		Id: *req.AnimeId,
	}
	dbRelAnimeUserLibrary, err := h.dbsRelAnimeUserLibrary.AddToLibrary(user, reqAnime)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: AddToLibrary: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"progress": dbRelAnimeUserLibrary,
	})
}

func (h *handlerUserLibraryAnime) GetProgress(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: GetProgress: GetUser: nil")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	opts := make([]database.OptionsFunc, 0)
	queries := r.URL.Query()
	relIdStr := queries.Get("rel_id")
	animeIdStr := queries.Get("anime_id")
	status := strings.ToLower(strings.TrimSpace(queries.Get("status")))

	opts = append(opts, database.WithStatus(status))
	err := uuid.Validate(relIdStr)
	if err == nil {
		id, err := uuid.Parse(relIdStr)
		if err == nil {
			opts = append(opts, database.WithRelId(id))
		}
	}
	err = uuid.Validate(animeIdStr)
	if err == nil {
		id, err := uuid.Parse(animeIdStr)
		if err == nil {
			opts = append(opts, database.WithAnimeId(id))
		}
	}

	progress, err := h.dbsRelAnimeUserLibrary.GetProgress(user, opts...)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: GetProgress: GetProgress: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"progress": progress,
	})
}

func (h *handlerUserLibraryAnime) SetProgress(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: GetUser: nil")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	type UpdateRequest struct {
		ProgressId *uuid.UUID `json:"progress_id"`
		Status     *string    `json:"status"`
		Episode    *int       `json:"episode"`
	}

	var req UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	if req.ProgressId == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing progress id")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	if req.Status == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing status")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}
	if req.Episode == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing episode")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	reqRelAnimeUserLibrary := &database.RelAnimeUserLibrary{
		Id:      *req.ProgressId,
		Status:  *req.Status,
		Episode: *req.Episode,
	}
	dbProgress, err := h.dbsRelAnimeUserLibrary.UpdateProgress(user, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: UpdateProgress: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"progress": dbProgress,
	})
}
