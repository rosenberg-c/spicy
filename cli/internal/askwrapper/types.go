package askwrapper

type HistoryEntry struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	At       int64  `json:"at,omitempty"`
}
