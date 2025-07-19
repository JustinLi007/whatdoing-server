package api

import (
	"github.com/JustinLi007/whatdoing-server/internal/database"
)

// TODO: maybe delete
type HandlerAnimeNames interface {
}

type handlerAnimeNames struct {
	dbsAnimeNames         database.DbsAnimeNames
	dbsRelAnimeAnimeNames database.DbsRelAnimeAnimeNames
}

var handlerAnimeNamesInstance *handlerAnimeNames

func NewHandlerAnimeNames(dbsAnimeNames database.DbsAnimeNames, dbsRelAnimeAnimeNames database.DbsRelAnimeAnimeNames) HandlerAnimeNames {
	if handlerAnimeNamesInstance != nil {
		return handlerAnimeNamesInstance
	}

	newHandlerAnimeNames := &handlerAnimeNames{
		dbsAnimeNames:         dbsAnimeNames,
		dbsRelAnimeAnimeNames: dbsRelAnimeAnimeNames,
	}
	handlerAnimeNamesInstance = newHandlerAnimeNames

	return handlerAnimeNamesInstance
}
