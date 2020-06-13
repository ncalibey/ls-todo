package models

// Todo is the model we use for encapsulating an individual todo. The tags you see are
// called "struct tags". They give metadata information that can help certain operations.
//
// In this case, the `json` tag is telling the JSON encoder/decor which json field maps to
// which field in the struct. The `db` tag is doing the same but for the database.
//
// These, strictly speaking, aren't necessary when there is a 1:1 match like we have here --
// the encoder/decoders are usually smart enough to figure this out. But they do allow us to
// specify different names if we want to (e.g. if the completed column in the db was "done" we
// could do `db:"done"` for the `Completed` field).
type Todo struct {
	ID          int64  `json:"id" db:"id"`
	Title       string `json:"title" db:"title"`
	Day         string `json:"day" db:"day"`
	Month       string `json:"month" db:"month"`
	Year        string `json:"year" db:"year"`
	Completed   bool   `json:"completed" db:"completed"`
	Description string `json:"description" db:"description"`
}
