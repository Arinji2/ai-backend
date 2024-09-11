package completions

import (
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
