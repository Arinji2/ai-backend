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

	if err := decoder.Decode(&requestBody); err != nil {
		return "", errors.New("invalid input")
	}

	var prompt string

	switch body := requestBody.(type) {
	case []interface{}:
		if len(body) > 0 {
			if bodyMap, ok := body[0].(map[string]interface{}); ok {
				if content, exists := bodyMap["content"].(string); exists {
					prompt = content
				} else {
					return "", errors.New("invalid input format: missing 'content' field")
				}
			} else if bodyString, ok := body[0].(string); ok {
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
