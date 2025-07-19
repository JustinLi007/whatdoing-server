package database

import "github.com/google/uuid"

func buildNamesMap(allNames []*RelAnimeAnimeNames) map[uuid.UUID][]*AnimeName {
	namesMap := make(map[uuid.UUID][]*AnimeName)

	for _, v := range allNames {
		curId := v.AnimeId

		_, ok := namesMap[curId]
		if !ok {
			namesMap[curId] = make([]*AnimeName, 0)
		}

		name := &AnimeName{
			Id:        v.AnimeName.Id,
			CreatedAt: v.AnimeName.CreatedAt,
			UpdatedAt: v.AnimeName.UpdatedAt,
			Name:      v.AnimeName.Name,
		}

		namesMap[curId] = append(namesMap[curId], name)
	}

	return namesMap
}
