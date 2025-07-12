package database

type DbsAnimeNames interface {
	GetNames() error
}

type PgDbsAnimeNames struct {
	db DbService
}

var dbsAnimeNamesInstance *PgDbsAnimeNames

func NewDbsAnimeNames(db DbService) DbsAnimeNames {
	if dbsAnimeNamesInstance != nil {
		return dbsAnimeNamesInstance
	}

	newDbsAnimeNames := &PgDbsAnimeNames{
		db: db,
	}
	dbsAnimeNamesInstance = newDbsAnimeNames

	return dbsAnimeNamesInstance
}

func (d *PgDbsAnimeNames) GetNames() error {
	return nil
}
