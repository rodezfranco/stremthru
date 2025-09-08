package meta_type

type IdType string

const (
	IdTypeMovie   IdType = "movie"
	IdTypeShow    IdType = "show"
	IdTypeUnknown IdType = ""
)

func (it IdType) IsValid() bool {
	return it == IdTypeUnknown || it == IdTypeMovie || it == IdTypeShow
}

type IdMapAnime struct {
	AniDB       string `json:"anidb,omitempty"`
	AniList     string `json:"anilist,omitempty"`
	AniSearch   string `json:"anisearch,omitempty"`
	AnimePlanet string `json:"animeplanet,omitempty"`
	Kitsu       string `json:"kitsu,omitempty"`
	LiveChart   string `json:"livechart,omitempty"`
	MAL         string `json:"mal,omitempty"`
	NotifyMoe   string `json:"notifymoe,omitempty"`
}

type IdMap struct {
	Type       IdType      `json:"type"`
	IMDB       string      `json:"imdb,omitempty"`
	TMDB       string      `json:"tmdb,omitempty"`
	TVDB       string      `json:"tvdb,omitempty"`
	TVMaze     string      `json:"tvmaze,omitempty"`
	Trakt      string      `json:"trakt,omitempty"`
	Letterboxd string      `json:"lboxd,omitempty"`
	Anime      *IdMapAnime `json:"anime,omitempty"`
}
