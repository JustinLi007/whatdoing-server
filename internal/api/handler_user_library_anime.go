package api

import (
	"database/sql"
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
	SetStatus(w http.ResponseWriter, r *http.Request)
	RemoveProgress(w http.ResponseWriter, r *http.Request)
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
		AnimeId *string `json:"anime_id"`
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
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	err = uuid.Validate(*req.AnimeId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: Validate: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	animeId, err := uuid.Parse(*req.AnimeId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: AddToLibrary: Parse: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	reqAnime := &database.Anime{
		Id: animeId,
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
	relIdStr := queries.Get("progress_id")
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
	if err == sql.ErrNoRows {
		progress = make([]*database.RelAnimeUserLibrary, 0)
	} else if err != nil {
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
		ProgressId *string `json:"progress_id"`
		Episode    *int    `json:"episode"`
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
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	err = uuid.Validate(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Validate: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	progressId, err := uuid.Parse(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Parse: %v", err)
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
		Id:      progressId,
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

func (h *handlerUserLibraryAnime) SetStatus(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: GetUser: nil")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	type UpdateStatusRequest struct {
		ProgressId *string `json:"progress_id"`
		Status     *string `json:"status"`
	}

	var req UpdateStatusRequest
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
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	err = uuid.Validate(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Validate: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	progressId, err := uuid.Parse(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Parse: %v", err)
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

	reqRelAnimeUserLibrary := &database.RelAnimeUserLibrary{
		Id:     progressId,
		Status: *req.Status,
	}
	dbProgress, err := h.dbsRelAnimeUserLibrary.UpdateStatus(user, reqRelAnimeUserLibrary)
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

func (h *handlerUserLibraryAnime) RemoveProgress(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: GetUser: nil")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	type UpdateRequest struct {
		ProgressId *string `json:"progress_id"`
	}

	var req UpdateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: Decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	if req.ProgressId == nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: missing progress id")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	err = uuid.Validate(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: Validate: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	progressId, err := uuid.Parse(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: Parse: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	reqRelAnimeUserLibrary := &database.RelAnimeUserLibrary{
		Id: progressId,
	}
	err = h.dbsRelAnimeUserLibrary.RemoveProgress(user, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: RemoveProgress: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{})
}
