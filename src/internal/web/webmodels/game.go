
package webmodels

type GameRequest struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type PlayerResponse struct {
	ID   string `json:"id"`
	Cell int    `json:"cell"`
}

type GameResponse struct {
	ID            string           `json:"id"`
	Board         [3][3]int        `json:"board"`
	Players       []PlayerResponse `json:"players"`
	CurrentPlayer string           `json:"current_player,omitempty"`
	Status        int               `json:"status"`
	Winner        string            `json:"winner,omitempty"`
	Mode          int               `json:"mode"`
}

type NewGameResponse struct {
	ID            string           `json:"id"`
	Board         [3][3]int        `json:"board"`
	Players       []PlayerResponse `json:"players"`
	CurrentPlayer string           `json:"current_player,omitempty"`
	Status        int               `json:"status"`
	Mode          int               `json:"mode"`
}
