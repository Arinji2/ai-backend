package input

import (
	"encoding/json"
	"errors"
	"net/http"
)

func CompletionsBodyInput(r *http.Request) (string, error) {
	var requestBody []interface{}
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&requestBody); err != nil {
		return "", errors.New("invalid input")

	}

	var prompt string
	if len(requestBody) > 0 {

		if bodyMap, ok := requestBody[0].(map[string]interface{}); ok {
			if content, exists := bodyMap["content"].(string); exists {
				prompt = content
			} else {
				return "", errors.New("invalid input format")
			}
		} else if bodyString, ok := requestBody[0].(string); ok {

			prompt = bodyString
		} else {
			return "", errors.New("invalid input format")
		}
	} else {

		return "", errors.New("empty input")
	}

	return prompt, nil
}
