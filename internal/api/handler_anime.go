package api

type HandlerAnime interface{}

type handlerAnime struct {
}

var handlerAnimeInstance *handlerAnime

func NewHandlerAnime() HandlerAnime {
	if handlerAnimeInstance != nil {
		return handlerAnimeInstance
	}

	newHandlerAnime := &handlerAnime{}
	handlerAnimeInstance = newHandlerAnime

	return handlerAnimeInstance
}
