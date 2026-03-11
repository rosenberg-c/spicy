package validator

type Action string

const (
	ActionContinue Action = "continue"
	ActionExit     Action = "exit"
)

type ValidationResponse struct {
	Action            Action   `json:"action"`
	Reason            string   `json:"reason"`
	SuggestedFilename string   `json:"suggested_filename,omitempty"`
	Suggestions       []string `json:"suggestions,omitempty"`
}

func (a Action) String() string {
	return string(a)
}

func (a Action) IsValid() bool {
	return a == ActionContinue || a == ActionExit
}
