package models

type Todo struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Day         string `json:"day"`
	Month       string `json:"month"`
	Year        string `json:"year"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
}
