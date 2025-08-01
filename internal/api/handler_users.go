package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type HandlerUsers interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	GetUserById(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	CheckSession(w http.ResponseWriter, r *http.Request)
}

type handlerUsers struct {
	dbsUsers database.DbsUsers
	dbsJwt   database.DbsJwt
}

var handlerUsersInstance *handlerUsers

func NewHandlerUsers(dbsUsers database.DbsUsers, dbsJwt database.DbsJwt) HandlerUsers {
	if handlerUsersInstance != nil {
		return handlerUsersInstance
	}

	newHandlerUsers := &handlerUsers{
		dbsUsers: dbsUsers,
		dbsJwt:   dbsJwt,
	}
	handlerUsersInstance = newHandlerUsers

	return handlerUsersInstance
}

func (h *handlerUsers) SignUp(w http.ResponseWriter, r *http.Request) {
	type ValidateRequest struct {
		Email    *string `json:"email"`
		Username *string `json:"username"`
		Password *string `json:"password"`
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

	newUser := &database.User{
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

	token, err := h.dbsJwt.Insert(createdUser.Id, time.Hour*12, time.Hour*24, tokens.ScopeAuthenticate)
	if err != nil {
		log.Printf("error: handler_users CreateUser dbs.CreateToken: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.SetCookie(w, "whatdoing-jwt", token.Token.PlainText)
	utils.SetCookie(w, "whatdoing-jwt-refresh", token.RefreshToken.PlainText)
	// FIX: change payload
	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"user": createdUser,
		"next": "/home",
	})
}

func (h *handlerUsers) GetUserById(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: handler users GetUserById: user is nil")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

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

	// TODO: idk what I want this handler for...
	if user.Id != userId {
		log.Printf("error: handler_users GetUserById user not equal requesting user")
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

func (h *handlerUsers) Login(w http.ResponseWriter, r *http.Request) {
	type ValidateRequest struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	var req ValidateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("error: handler_users Login Decode: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "failed to decode request body.",
		})
		return
	}

	if req.Email == nil {
		log.Printf("error: handler_users Login: no email")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "missing required credentials",
		})
		return
	}
	if req.Password == nil {
		log.Printf("error: handler_users Login: no password")
		utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "missing required credentials",
		})
		return
	}

	newUser := &database.User{
		Email: *req.Email,
	}

	err = newUser.Password.Set(*req.Password)
	if err != nil {
		log.Printf("error: handler_users Login Set password: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	existingUser, err := h.dbsUsers.GetUserByEmailPassword(newUser)
	if err != nil {
		log.Printf("error: handler_users Login dbs.GetUserByEmailPassword: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	token, err := h.dbsJwt.Insert(existingUser.Id, time.Hour*12, time.Hour*24, tokens.ScopeAuthenticate)
	if err != nil {
		log.Printf("error: handler_users Login dbs.Insert: %v", err)
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.SetCookie(w, "whatdoing-jwt", token.Token.PlainText)
	utils.SetCookie(w, "whatdoing-jwt-refresh", token.RefreshToken.PlainText)
	// FIX: change payload
	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"user": existingUser,
		"next": "/home",
	})
}
func (h *handlerUsers) CheckSession(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: handler_users CheckSession: user nil")
		utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		return
	}

	utils.WriteJson(w, http.StatusOK, utils.Envelope{
		"user": user,
	})
}
