package api

import (
	"log"
	"net/http"

	"github.com/JustinLi007/whatdoing-server/internal/database"
	"github.com/JustinLi007/whatdoing-server/internal/utils"
)

type HandlerJwt interface {
	CreateToken(w http.ResponseWriter, r *http.Request)
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

func (h *handlerJwt) CreateToken(w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJson(w, http.StatusOK, utils.Envelope{})
	log.Printf("error: handler jwt CreateToken: Write Json: %v", err)
}
