package userword

type AddWordRequest struct {
	WordID  *uint64 `json:"wordId"`
	Word    *string `json:"word"`
	Meaning *string `json:"meaning"`
}

type ReviewRequest struct {
	Quality int `json:"quality" binding:"min=0,max=5"`
}
