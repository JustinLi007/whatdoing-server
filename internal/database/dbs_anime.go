package database

type DbsAnime interface {
}

type PgDbsAnime struct {
}

var dbsAnimeInstance *PgDbsAnime

func NewDbsAnime() DbsAnime {
	if dbsAnimeInstance != nil {
		return dbsAnimeInstance
	}

	newDbsAnime := &PgDbsAnime{}
	dbsAnimeInstance = newDbsAnime

	return dbsAnimeInstance
}
