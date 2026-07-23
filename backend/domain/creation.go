package domain

type TextReq struct {
	Text   string `json:"text" validate:"required"`
	Action string `json:"action"` // action: improve, summary, extend, shorten, etc.
}

// FIM (Fill in Middle) tokens
const (
	FIMPrefix = "<fim_prefix>"
	FIMSuffix = "<fim_suffix>"
	FIMMiddle = "<fim_middle>"
)

type CompleteReq struct {
	// For FIM (Fill in Middle) style completion
	Prefix string `json:"prefix,omitempty"`
	Suffix string `json:"suffix,omitempty"`
}
