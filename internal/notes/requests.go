package notes

type CreateNoteRequest struct {
	Title     string `json:"title" validate:"required"`
	Content   string `json:"content"`
	IsPrivate bool   `json:"is_private"`
}

type UpdateNoteRequest struct {
	Title     *string `json:"title"`
	Content   *string `json:"content"`
	IsPrivate *bool   `json:"is_private"`
}
