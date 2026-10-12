package userword

type AddWordRequest struct {
	WordID  *uint64 `json:"wordId"`
	Word    *string `json:"word"`
	Meaning *string `json:"meaning"`
}

// Quality is a pointer so a missing value is rejected instead of being read
// as 0, the worst grade, which would reset the word's progress.
type ReviewRequest struct {
	Quality *int `json:"quality" binding:"required,min=0,max=5"`
}
