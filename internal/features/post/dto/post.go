package dto

import "time"

// PostResponse represents the public post data.
type PostResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  string    `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePostRequest represents the payload for creating a post.
type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=3"`
	Content string `json:"content" validate:"required,min=10"`
}

// UpdatePostRequest represents the payload for updating a post.
type UpdatePostRequest struct {
	Title   string `json:"title,omitempty" validate:"omitempty,min=3"`
	Content string `json:"content,omitempty" validate:"omitempty,min=10"`
}
