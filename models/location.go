package models

type Location struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
}
