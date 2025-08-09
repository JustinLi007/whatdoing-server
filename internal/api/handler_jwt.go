package api

import (
	"log"
	"net/http"
	"time"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/tokens"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
)

type HandlerJwt interface {
	RefreshJwt(w http.ResponseWriter, r *http.Request)
}

type handlerJwt struct {
	dbsJwt database.DbsJwt
}

var handlerJwtInstance *handlerJwt

func NewHandlerJwt(dbs database.DbsJwt) HandlerJwt {
	if handlerJwtInstance != nil {
		return handlerJwtInstance
	}

	newHandlerJwt := &handlerJwt{
		dbsJwt: dbs,
	}
	handlerJwtInstance = newHandlerJwt

	return handlerJwtInstance
}

func (h *handlerJwt) RefreshJwt(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUser(r)
	if user == nil {
		log.Printf("error: Handler: Jwt: RefreshJwt: GetUser: user nil")
		err := utils.WriteJson(w, http.StatusBadRequest, utils.Envelope{
			"error": "bad request",
		})
		log.Printf("error: Handler: Jwt: RefreshJwt: GetUser: WriteJson: %v", err)
		return
	}

	dbJwt, err := h.dbsJwt.Get(user)
	if err != nil {
		log.Printf("error: Handler: Jwt: RefreshJwt: Get: %v", err)
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Jwt: RefreshJwt: Get: WriteJson: %v", err)
		return
	}

	expired := time.Now().After(dbJwt.RefreshToken.Expiry)
	if expired {
		log.Printf("error: Handler: Jwt: RefreshJwt: token expired")
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Jwt: RefreshJwt: token expired: WriteJson: %v", err)
		return
	}

	newJwt, err := h.dbsJwt.Insert(user.Id, time.Hour*12, time.Hour*24, tokens.ScopeAuthenticate)
	if err != nil {
		log.Printf("error: Handler: Jwt: RefreshJwt: Insert: %v", err)
		err := utils.WriteJson(w, http.StatusInternalServerError, utils.Envelope{
			"error": "internal server error",
		})
		log.Printf("error: Handler: Jwt: RefreshJwt: Insert: WriteJson: %v", err)
		return
	}

	utils.SetCookie(w, "whatdoing-jwt", newJwt.Token.PlainText)
	utils.SetCookie(w, "whatdoing-jwt-refresh", newJwt.RefreshToken.PlainText)
	err = utils.WriteJson(w, http.StatusOK, utils.Envelope{})
	if err != nil {
		log.Printf("error: Handler: Jwt: RefreshJwt: payload: WriteJson: %v", err)
	}
}
