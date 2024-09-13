package tasks

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	custom_log "github.com/Arinji2/ai-backend/logger"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func GetModel(task *TaskObject) (*genai.GenerativeModel, context.Context) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(task.ApiKey))
	if err != nil {
		panic(err)
	}
	model := client.GenerativeModel("gemini-1.5-flash")
	model.SetTemperature(2.0)
	return model, ctx
}

func GetPromptResponse(task *TaskObject, prompt string) (string, error) {
	model, context := GetModel(task)
	resp, err := model.GenerateContent(context, genai.Text(prompt))
	if err != nil {

		return "", err
	}

	parsed := slices.Concat(resp.Candidates[0].Content.Parts)
	stringPrompt := ""
	for _, part := range parsed {
		stringPrompt = fmt.Sprintf(`%s %s`, stringPrompt, part)
	}
	re := regexp.MustCompile(`[~#@\$%\^&\*\+=\-\_/|\\"]`)

	result := re.ReplaceAllString(stringPrompt, "")
	result = strings.TrimSpace(strings.ReplaceAll(result, "\n", " "))
	return result, nil

}

func GetModelStatus(task *TaskObject) bool {
	model, context := GetModel(task)
	resp, err := model.GenerateContent(context, genai.Text(`Respond with only the word "testing" without the quotes.`))
	if err != nil {

		return false
	}

	parsed := slices.Concat(resp.Candidates[0].Content.Parts)
	stringPrompt := ""
	for _, part := range parsed {
		stringPrompt = fmt.Sprintf(`%s %s`, stringPrompt, part)
	}
	re := regexp.MustCompile(`[~#@\$%\^&\*\+=\-\_/|\\"]`)

	result := re.ReplaceAllString(stringPrompt, "")
	result = strings.TrimSpace(strings.ReplaceAll(result, "\n", " "))
	if result == "testing" {
		return true
	}
	custom_log.Logger.Debug("Issue with Content", result)
	return true

}
