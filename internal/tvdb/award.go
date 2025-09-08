package tvdb

type Award struct {
	Id        int    `json:"id"`
	Year      string `json:"year"`
	Details   any    `json:"details"`
	IsWinner  bool   `json:"isWinner"`
	Category  string `json:"category"`
	Name      string `json:"name"`
	Series    any    `json:"series"`
	Movie     any    `json:"movie"`
	Episode   any    `json:"episode"`
	Character any    `json:"character"`
}

type Awards []Award
