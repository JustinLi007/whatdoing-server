package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/google/uuid"
)

type HandlerAnimeAltNames interface {
	AddAltName(w http.ResponseWriter, r *http.Request)
	DeleteAltNames(w http.ResponseWriter, r *http.Request)
}

type handlerAnimeAltNames struct {
	dbsAnimeAltNames database.DbsAnimeAltNames
}

var handlerAnimeAltNamesInstance *handlerAnimeAltNames

func NewHandlerAnimeAltNames(dbsAnimeAltNames database.DbsAnimeAltNames) HandlerAnimeAltNames {
	if handlerAnimeAltNamesInstance != nil {
		return handlerAnimeAltNamesInstance
	}

	newHandlerAnimeAltNames := &handlerAnimeAltNames{
		dbsAnimeAltNames: dbsAnimeAltNames,
	}
	handlerAnimeAltNamesInstance = newHandlerAnimeAltNames

	return handlerAnimeAltNamesInstance
}

func (h *handlerAnimeAltNames) AddAltName(w http.ResponseWriter, r *http.Request) {
	type AddAltNameRequest struct {
		AnimeId         *string `json:"anime_id"`
		AlternativeName *string `json:"alternative_name"`
	}

	req := AddAltNameRequest{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Decode: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Decode: WriteJson: %v", err)
		}
		return
	}

	if req.AnimeId == nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Missing Anime Id")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Missing Anime Id: WriteJson: %v", err)
		}
		return
	}

	if req.AlternativeName == nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Missing Alt Name")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Missing Alt Name: WriteJson: %v", err)
		}
		return
	}

	newAltName := strings.TrimSpace(*req.AlternativeName)
	if newAltName == "" {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Alt Name Blank")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Alt Name Blank: WriteJson: %v", err)
		}
		return
	}

	if err := uuid.Validate(*req.AnimeId); err != nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Validate: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Validate: WriteJson: %v", err)
		}
		return
	}

	animeId, err := uuid.Parse(*req.AnimeId)
	if err != nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Parse: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: Parse: WriteJson: %v", err)
		}
		return
	}

	reqAnimeAltName := database.AnimeAltName{
		AnimeId: animeId,
		AnimeName: database.AnimeName{
			Name: newAltName,
		},
	}
	if err := h.dbsAnimeAltNames.AddAltName(&reqAnimeAltName); err != nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: AddAltName: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: AddAltName: AddAltName: WriteJson: %v", err)
		}
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, utils.Envelope{}); err != nil {
		log.Printf("error: Handler: AnimeAltNames: AddAltName: Decode: WriteJson: %v", err)
	}
}

func (h *handlerAnimeAltNames) DeleteAltNames(w http.ResponseWriter, r *http.Request) {
	type DeleteAltNamesRequest struct {
		AnimeId       *string  `json:"anime_id"`
		AnimeNamesIds []string `json:"anime_names_ids"`
	}

	var req DeleteAltNamesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Decode: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Decode: WriteJson: %v", err)
		}
		return
	}

	if req.AnimeId == nil {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Missing Anime Id")
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Missing Anime Id: WriteJson: %v", err)
		}
		return
	}

	if err := uuid.Validate(*req.AnimeId); err != nil {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Validate: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Validate: WriteJson: %v", err)
		}
		return
	}

	animeId, err := uuid.Parse(*req.AnimeId)
	if err != nil {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Parse: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: Parse: WriteJson: %v", err)
		}
		return
	}

	badIds := make([]string, 0)
	reqAltNames := make([]*database.AnimeAltName, 0)
	for _, v := range req.AnimeNamesIds {
		if err := uuid.Validate(v); err != nil {
			badIds = append(badIds, v)
			continue
		}

		id, err := uuid.Parse(v)
		if err != nil {
			badIds = append(badIds, v)
			continue
		}

		reqAltNames = append(reqAltNames, &database.AnimeAltName{
			AnimeId: animeId,
			AnimeName: database.AnimeName{
				Id: id,
			},
		})
	}

	if err := h.dbsAnimeAltNames.DeleteAltNames(reqAltNames); err != nil && err != sql.ErrNoRows {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: DeleteAltNames: %v", err)
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: DeleteAltNames: WriteJson: %v", err)
		}
		return
	}

	if len(badIds) > 0 {
		if err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": fmt.Sprintf("failed to delete the following alternative names: %v", strings.Join(badIds, ", ")),
		}); err != nil {
			log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: DeleteAltNames: WriteJson: %v", err)
		}
		return
	}

	if err := utils.WriteJson(w, http.StatusOK, utils.Envelope{}); err != nil {
		log.Printf("error: Handler: AnimeAltNames: DeleteAltNames: DeleteAltNames: WriteJson: %v", err)
	}
}
