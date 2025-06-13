package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type HandlerUsers interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
}

type handlerUsers struct {
	dbsUsers database.DbsUsers
}

var handlerInstance *handlerUsers

func NewHandlerUsers(dbs database.DbsUsers) HandlerUsers {
	if handlerInstance != nil {
		return handlerInstance
	}

	newHandlerUsers := &handlerUsers{
		dbsUsers: dbs,
	}
	handlerInstance = newHandlerUsers

	return handlerInstance
}

func (h *handlerUsers) CreateUser(w http.ResponseWriter, r *http.Request) {
	type ValidateRequest struct {
		Email    *string
		Username *string
		Password *string
	}

	var req ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: handler_users CreateUser Decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "failed to decode request body.",
		})
		return
	}

	if req.Email == nil {
		log.Printf("error: handler_users CreateUser: no email")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "missing required credentials",
		})
		return
	}
	if req.Password == nil {
		log.Printf("error: handler_users CreateUser: no password")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "missing required credentials",
		})
		return
	}

	newUser := database.User{
		Email:    *req.Email,
		Username: req.Username,
	}

	err = newUser.Password.Set(*req.Password)
	if err != nil {
		log.Printf("error: handler_users CreateUser Set password: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	createdUser, err := h.dbsUsers.CreateUser(newUser)
	if err != nil {
		log.Printf("error: handler_users CreateUser dbs.CreateUser: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.SetCookie(w, "whatdoing-test-cookie", "fk you")

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"user": createdUser,
	})
}

func (h *handlerUsers) GetUserById(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "userId")

	err := uuid.Validate(userIdStr)
	if err != nil {
		log.Printf("error: handler_users GetUserById validate userIdStr: %v", err)
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		return
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		log.Printf("error: handler_users GetUserById parse userIdStr: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	existingUser, err := h.dbsUsers.GetUserById(userId)
	if err != nil {
		log.Printf("error: handler_users GetUserById dbs.GetUserById: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"user": existingUser,
	})
}
