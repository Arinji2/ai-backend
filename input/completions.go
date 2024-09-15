package input

import (
	"encoding/json"
	"errors"
	"net/http"
)

func CompletionsBodyInput(r *http.Request) (string, error) {
	var requestBody interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	// Decode the request body into an interface{}
	if err := decoder.Decode(&requestBody); err != nil {
		return "", errors.New("invalid input")
	}

	var prompt string

	switch body := requestBody.(type) {
	// Handle the case where input is an array of objects (with "role" and "content")
	case []interface{}:
		if len(body) > 0 {
			// Check if the first item is an object with "role" and "content"
			if bodyMap, ok := body[0].(map[string]interface{}); ok {
				if content, exists := bodyMap["content"].(string); exists {
					prompt = content
				} else {
					return "", errors.New("invalid input format: missing 'content' field")
				}
			} else if bodyString, ok := body[0].(string); ok {
				// Handle the case where the input is an array of strings
				prompt = bodyString
			} else {
				return "", errors.New("invalid input format for array")
			}
		} else {
			return "", errors.New("empty input array")
		}
	default:
		return "", errors.New("unknown input format")
	}

	return prompt, nil
}
