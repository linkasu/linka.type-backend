package data

// CreateStatementRequest структура для создания statement
type CreateStatementRequest struct {
	Title      string `json:"title" binding:"required"`
	CategoryID string `json:"categoryId" binding:"required"`
}

// UpdateStatementRequest структура для обновления statement
type UpdateStatementRequest struct {
	Title      string `json:"title" binding:"required"`
	CategoryID string `json:"categoryId" binding:"required"`
}

// CreateCategoryRequest структура для создания категории
type CreateCategoryRequest struct {
	Title string `json:"title" binding:"required"`
}

// UpdateCategoryRequest структура для обновления категории
type UpdateCategoryRequest struct {
	Title string `json:"title" binding:"required"`
}