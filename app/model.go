package app

import "github.com/google/uuid"

type Risk struct {
	ID          uuid.UUID `json:"id"`
	State       string    `json:"state"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
}

type CreateRisk struct {
	State       string `json:"state" binding:"required,oneof=open closed accepted investigating"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
