package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// StatusMessage is a generic JSON response containing the message for the HTTP status.
type StatusMessage struct {
	Message string
}

// Health is a JSON response for healthcheck request
type Health struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ValidationMessage is a function to convert Gin-Gonic validation errors to `StatusMessage`.
func ValidationMessage(err error) StatusMessage {
	return StatusMessage{
		Message: strings.Join(ValidationMessages(err), " "),
	}
}

// ValidationMessages is a function to convert Gin-Gonic validation error to human readable.
func ValidationMessages(err error) []string {
	if ve, ok := err.(validator.ValidationErrors); ok {
		out := make([]string, len(ve))
		for i, fe := range ve {
			out[i] = fe.Error()
		}
		return out
	} else if je, ok := err.(*json.UnmarshalTypeError); ok {
		return []string{fmt.Sprintf("The field %s must be a %s", je.Field, je.Type.String())}
	}
	return nil
}

// RecaptchaResponse is a JSON response from Google ReCaptcha.
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}
