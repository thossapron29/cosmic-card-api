package draws

type RevealDrawRequest struct {
	UserID          string `json:"userId"`
	DeckID          int64  `json:"deckId"`
	DrawMode        string `json:"drawMode"`
	Locale          string `json:"locale"`
	QuestionText    string `json:"questionText"`
	ClientLocalDate string `json:"clientLocalDate"`
}

type RevealDrawResponse struct {
	DrawID int64    `json:"drawId"`
	Card   DrawCard `json:"card"`
	Deck   DrawDeck `json:"deck"`
}

type DrawCard struct {
	ID               int64  `json:"id"`
	Code             string `json:"code"`
	Title            string `json:"title"`
	ShortMessage     string `json:"shortMessage"`
	Meaning          string `json:"meaning"`
	ReflectionPrompt string `json:"reflectionPrompt"`
	ShareText        string `json:"shareText"`
	IllustrationKey  string `json:"illustrationKey"`
	EnergyType       string `json:"energyType"`
}

type DrawDeck struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}
