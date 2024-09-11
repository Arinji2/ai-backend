package completions

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func CompletionsHandler(w http.ResponseWriter, r *http.Request) {

	prompt, err := GetBodyInput(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Println(prompt)

	w.Write([]byte("Success"))

}

func GetBodyInput(r *http.Request) (string, error) {
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

	fmt.Println("Prompt:", prompt)
	return prompt, nil
}
