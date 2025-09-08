package stremio_shared

func GetCinemetaPosterURL(imdbId string) string {
	return "https://images.metahub.space/poster/small/" + imdbId + "/img"
}

func GetCinemetaBackgroundURL(imdbId string) string {
	return "https://images.metahub.space/background/medium/" + imdbId + "/img"
}
