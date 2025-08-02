package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
)

type HandlerUsers interface {
	SignUp(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	CheckSession(w http.ResponseWriter, r *http.Request)
	Logout(w http.ResponseWriter, r *http.Request)
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

func (h *handlerUsers) Logout(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: Users: Logout: GetUser: user nil")
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Users: Logout: GetUser: WriteJson: %v", err)
		return
	}

	cookie, err := r.Cookie("whatdoing-jwt")
	if err != nil {
		log.Printf("error: Handler: Users: Logout: Cookie: whatdoing-jwt: cookie nil")
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Users: Logout: Cookie: WriteJson: %v", err)
		return
	}

	err = cookie.Valid()
	if err != nil {
		log.Printf("error: Handler: Users: Logout: Cookie: Valid: %v", err)
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Users: Logout: Cookie: Valid: WriteJson: %v", err)
		return
	}

	reqJwt := &tokens.Jwt{
		Token: &tokens.Token{
			PlainText: cookie.Value,
		},
	}
	err = h.dbsJwt.Delete(user, reqJwt)
	if err != nil {
		log.Printf("error: Handler: Users: Logout: Delete: %v", err)
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Users: Logout: Delete WriteJson: %v", err)
		return
	}

	utils.DeleteCookie(w, "whatdoing-jwt")
	utils.DeleteCookie(w, "whatdoing-jwt-refresh")
	err = utils.WriteJson(w, http.StatusOK, utils.Envelope{})
	if err != nil {
		log.Printf("error: Handler: Users: Logout: payload: WriteJson: %v", err)
	}
}
