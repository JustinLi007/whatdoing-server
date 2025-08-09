package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerProgressAnime interface {
	AddToLibrary(w http.ResponseWriter, r *http.Request)
	GetProgress(w http.ResponseWriter, r *http.Request)
	SetProgress(w http.ResponseWriter, r *http.Request)
	RemoveProgress(w http.ResponseWriter, r *http.Request)
}

type handlerProgressAnime struct {
	dbsUserLibrary   database.DbsUserLibrary
	dbsProgressAnime database.DbsProgressAnime
}

var handlerProgressAnimeInstance *handlerProgressAnime

func NewHandlerProgressAnime(dbsUserLibrary database.DbsUserLibrary, dbsRelAnimeUserLibrary database.DbsProgressAnime) HandlerProgressAnime {
	if handlerProgressAnimeInstance != nil {
		return handlerProgressAnimeInstance
	}
	newHandlerProgressAnime := &handlerProgressAnime{
		dbsUserLibrary:   dbsUserLibrary,
		dbsProgressAnime: dbsRelAnimeUserLibrary,
	}
	handlerProgressAnimeInstance = newHandlerProgressAnime

	return handlerProgressAnimeInstance
}

func (h *handlerProgressAnime) AddToLibrary(w http.ResponseWriter, r *http.Request) {
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
	dbRelAnimeUserLibrary, err := h.dbsProgressAnime.AddToLibrary(user, reqAnime)
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

func (h *handlerProgressAnime) GetProgress(w http.ResponseWriter, r *http.Request) {
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

	status := queries.Get("status")
	search := queries.Get("search")
	sort := queries.Get("sort")
	opts = append(opts, database.WithStatus(status))
	opts = append(opts, database.WithSearch(search))
	opts = append(opts, database.WithSort(sort))

	progressIdStr := queries.Get("progress_id")
	animeIdStr := queries.Get("anime_id")
	err := uuid.Validate(progressIdStr)
	if err == nil {
		id, err := uuid.Parse(progressIdStr)
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

	progress, err := h.dbsProgressAnime.GetProgress(user, opts...)
	if err == sql.ErrNoRows {
		progress = make([]*database.ProgressAnime, 0)
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

func (h *handlerProgressAnime) SetProgress(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: GetUser: nil")
		if err := utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: GetUser: WriteJson: %v", err)
		}
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
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: Decode: WriteJson: %v", err)
		}
		return
	}

	if req.ProgressId == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing progress id")
		if err := utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing progress id: WriteJson: %v", err)
		}
		return
	}
	err = uuid.Validate(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Validate: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: Validate: WriteJson: %v", err)
		}
		return
	}
	progressId, err := uuid.Parse(*req.ProgressId)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: Parse: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: Parse: WriteJson: %v", err)
		}
		return
	}

	if req.Episode == nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing episode")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: missing episode: WriteJson: %v", err)
		}
		return
	}

	reqRelAnimeUserLibrary := &database.ProgressAnime{
		Id:      progressId,
		Episode: *req.Episode,
	}
	_, err = h.dbsProgressAnime.UpdateProgress(user, reqRelAnimeUserLibrary)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: UpdateProgress: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: UserLibraryAnime: SetProgress: UpdateProgress: WriteJson: %v", err)
		}
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, utils.Envelope{}); err != nil {
		log.Printf("error: Handler: UserLibraryAnime: SetProgress: payload: WriteJson: %v", err)
	}
}

func (h *handlerProgressAnime) RemoveProgress(w http.ResponseWriter, r *http.Request) {
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

	reqRelAnimeUserLibrary := &database.ProgressAnime{
		Id: progressId,
	}
	err = h.dbsProgressAnime.RemoveProgress(user, reqRelAnimeUserLibrary)
	if err != nil {
		log.Printf("error: Handler: UserLibraryAnime: RemoveProgress: RemoveProgress: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{})
}
