package decks

type Deck struct {
	ID               int64  `json:"id"`
	Code             string `json:"code"`
	Name             string `json:"name"`
	ShortDescription string `json:"shortDescription"`
	CoverImage       string `json:"coverImage"`
	IconName         string `json:"iconName"`
	IsPremium        bool   `json:"isPremium"`
}
