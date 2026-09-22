package word

type CreateWordRequest struct {
	Word    string `json:"word" binding:"required"`
	Meaning string `json:"meaning" binding:"required"`

	Global bool `json:"global"`
}

type UpdateWordRequest struct {
	Word    *string `json:"word"`
	Meaning *string `json:"meaning"`
}
