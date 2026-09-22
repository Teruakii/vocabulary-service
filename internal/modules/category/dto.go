package category

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CategoryView struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	IsGlobal  bool   `json:"isGlobal"`
	WordCount int    `json:"wordCount"`
	MaxWords  int    `json:"maxWords"`
}
