package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type ImageRequest struct {
	Image    string `json:"image"`    // base64 encoded image string
	ImageType string `json:"imageType"` // e.g., "image/jpeg", "image/png"
}

// type NovaResponse struct {
// 	Output     NovaOutput `json:"output"`
// 	StopReason string     `json:"stopReason"`
// 	Usage      NovaUsage  `json:"usage"`
// }

// type NovaOutput struct {
// 	Message NovaMessage `json:"message"`
// }

// type NovaMessage struct {
// 	Content []NovaContent `json:"content"`
// 	Role    string        `json:"role"`
// }

// type NovaContent struct {
// 	Text string `json:"text"`
// }

// type NovaUsage struct {
// 	InputTokens               int `json:"inputTokens"`
// 	OutputTokens              int `json:"outputTokens"`
// 	TotalTokens               int `json:"totalTokens"`
// 	CacheReadInputTokenCount  int `json:"cacheReadInputTokenCount"`
// 	CacheWriteInputTokenCount int `json:"cacheWriteInputTokenCount"`
// }

// type ProcessedResponse struct {
// 	Success        bool        `json:"success"`
// 	Data           interface{} `json:"data"`
// 	Operation      string      `json:"operation,omitempty"`
// 	OriginalCount  int         `json:"originalCount,omitempty"`
// 	ProcessedCount int         `json:"processedCount,omitempty"`
// 	Error          string      `json:"error,omitempty"`
// }

func HandleDataFromPDF(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var imageRequestData ImageRequest

		if err := e.BindBody(&imageRequestData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		if imageRequestData.Image == "" {
			return e.BadRequestError("No image provided", nil)
		}

		processedData, err := process_image(imageRequestData.Image, imageRequestData.ImageType)
		if err != nil {
			fmt.Print(err.Error())
			return e.JSON(http.StatusInternalServerError, ProcessedResponse{
				Success: false,
				Error:   err.Error(),
				Data:    nil,
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
			"data":   processedData,
		})
	}
}

func process_image(base64Image string, mediaType string) (interface{}, error) {
	if mediaType == "" {
    mediaType = "jpeg"
}

// Load AWS configuration from environment variables
cfg, err := config.LoadDefaultConfig(
    context.TODO(),
    config.WithRegion("us-east-1"),
)
if err != nil {
    return nil, fmt.Errorf("failed to load AWS config: %v", err)
}

client := bedrockruntime.NewFromConfig(cfg)

systemList := []map[string]interface{}{{ 
	"text": ``,
},}

// Use the direct Nova model format with schemaVersion
messageList := []map[string]interface{}{
	{
		"role": "user",
		"content": []map[string]interface{}{
			{
				"image": map[string]interface{}{
					"format": mediaType,
					"source": map[string]interface{}{
						"bytes": base64Image,
					},
				},
			},
			{
				"text": `This is a recipe, please give me the following in the form of a valid json object:
						1. a list of ingredients with quantities and units
						2. a list of step-by-step instructions
						3. a title for the recipe
						4. a brief description of the recipe if available
						5. estimated total time
						6. number of servings
						7. cuisine type (e.g., Italian, Chinese, Mexican, etc.)
						8. meal type (e.g., breakfast, lunch, dinner, snack, dessert)
						9. country of origin
						10. an image of the final dish if available
						return exactly what it says on the page or "unknown" if not available.`,
			},
		},
	},
}

infParams := map[string]interface{}{
	"maxTokens":   4000,
	"temperature": 0.7,
	"topP":        0.9,
	"topK":        40,
}
	
requestBody := map[string]interface{}{
	"schemaVersion": 	"messages-v1",
	"messages":      	messageList,
	"system":        	systemList,
	"inferenceConfig": 	infParams,
}

jsonBody, err := json.Marshal(requestBody)
if err != nil {
    return nil, fmt.Errorf("failed to marshal request body: %v", err)
}

bedrockInput := &bedrockruntime.InvokeModelInput{
    ModelId:     aws.String("us.amazon.nova-lite-v1:0"),
    ContentType: aws.String("application/json"),
    Accept:      aws.String("application/json"),
    Body:        jsonBody,
}

output, err := client.InvokeModel(context.TODO(), bedrockInput)
if err != nil {
    return nil, fmt.Errorf("failed to invoke model: %v", err)
}

var response NovaResponse
if err := json.NewDecoder(bytes.NewReader(output.Body)).Decode(&response); err != nil {
    return nil, fmt.Errorf("failed to decode response: %v", err)
}
fmt.Print(response);
aiResponse := response.Output.Message.Content[0].Text
fmt.Printf("\n\n\nAI Response: %s\n", aiResponse)
// Clean the JSON response
cleanJson := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(aiResponse, "```json", ""), "```", ""))

var processedData interface{}
if err := json.Unmarshal([]byte(extractInner(cleanJson)), &processedData); err != nil {
    return nil, fmt.Errorf("invalid JSON response from AI: %v", err)
}

return processedData, nil
}