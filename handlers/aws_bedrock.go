package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type GroceryListRequest struct {
	List string `json:"list"`
}

type NovaResponse struct {
	Output     NovaOutput `json:"output"`
	StopReason string     `json:"stopReason"`
	Usage      NovaUsage  `json:"usage"`
}

// NovaOutput contains the message response
type NovaOutput struct {
	Message NovaMessage `json:"message"`
}

// NovaMessage represents the assistant's message
type NovaMessage struct {
	Content []NovaContent `json:"content"`
	Role    string        `json:"role"`
}

// NovaContent represents individual content blocks
type NovaContent struct {
	Text string `json:"text"`
}

// NovaUsage contains token usage information
type NovaUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	TotalTokens              int `json:"totalTokens"`
	CacheReadInputTokenCount int `json:"cacheReadInputTokenCount"`
	CacheWriteInputTokenCount int `json:"cacheWriteInputTokenCount"`
}

type ProcessedResponse struct {
	Success        bool        `json:"success"`
	Data           interface{} `json:"data"`
	Operation      string      `json:"operation,omitempty"`
	OriginalCount  int         `json:"originalCount,omitempty"`
	ProcessedCount int         `json:"processedCount,omitempty"`
	Error          string      `json:"error,omitempty"`
}

func HandleGroceryList(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var groceryListData GroceryListRequest

		if err := e.BindBody(&groceryListData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}
		// add bedrock request handling logic here
		processedData, err := processWithBedrock(groceryListData.List)
		if err != nil {
			return e.JSON(http.StatusInternalServerError, ProcessedResponse{
				Success: false,
				Error:   err.Error(),
				Data:    nil,
			})
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
			"result": processedData,
			"data": groceryListData,
		})
	}
}

func processWithBedrock(input string) (interface{}, error) {
	// Load AWS configuration from environment variables
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion("us-east-2"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := bedrockruntime.NewFromConfig(cfg)

	systemPrompt := `this is a list of recipe ingredients with quantities, unit, name, and possibly how it should be prepped.  i want you to give me a grocery list for the items i should buy to be able to make the recipe.  a couple things to note include:
		1. combine the ingredients names are the same, cobining quantities and units appropriatly
		2. only give me a json encoded list, no other text
		3. make sure the return object has the same structure as the input data
		4. example: 1 tsp lemon zest and 1 tbsp lemon juice would combine to 1 lemon`
		// 5. group the resulting list by grocery section: Produce, Meat, Dry Goods, Snacks/Beverages, Dairy`

	requestBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						// "text": "write me a haiku", // Combine system + user
						"text": systemPrompt + "\n\n" + input, // Combine system + user
					},
				},
			},
		},
		"inferenceConfig": map[string]interface{}{
			"maxTokens":   4000,
			"temperature": 0.1,
		},
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	bedrockInput := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("us.amazon.nova-micro-v1:0"), // meta.llama3-3-70b-instruct-v1:0 | us.amazon.nova-lite-v1:0 | us.amazon.nova-micro-v1:0 | amazon.nova-micro-v1:0 | amazon.nova-premier-v1:0
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
	// if response.Output ==  {
	// 	return nil, fmt.Errorf("empty response from AI model")
	// }
	// cost, err := calc_pricing(response.Usage)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to calculate pricing: %v", err)
	// }
	// fmt.Printf("\n\n\nAI usage (tokens):\ninput: %+v\noutput: %+v\ntotal: %+v\ncost: %+v\n\n\n", response.Usage.InputTokens, response.Usage.OutputTokens, response.Usage.TotalTokens, cost)
	aiResponse := response.Output.Message.Content[0].Text

	// Clean the JSON response
	cleanJson := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(aiResponse, "```json", ""), "```", ""))
	var processedData interface{}
	if err := json.Unmarshal([]byte(extractInner(cleanJson)), &processedData); err != nil {
		return nil, fmt.Errorf("invalid JSON response from AI: %v", err)
	}
	// fmt.Println("Processed Data:", processedData)
	return processedData, nil
}

func extractInner(s string) string {
	start := strings.Index(s, "[")
	end := strings.LastIndex(s, "]")
	
	if start == -1 || end == -1 || end < start {
		return ""
	}
	
	return s[start:end+1]
}

// func calc_pricing(usage NovaUsage) (float64, error) {
// 	input_price := 0.000035/1000
// 	output_price := 0.00014/1000
// 	if usage.InputTokens < 0 || usage.OutputTokens < 0 {
// 		return 0, fmt.Errorf("invalid token counts: input %d, output %d", usage.InputTokens, usage.OutputTokens)
// 	}

// 	total_cost := float64(usage.InputTokens)*input_price + float64(usage.OutputTokens)*output_price
// 	total_cost = float64(int(total_cost*10000*100)) / 100 // multiply by 10000 and round to 2 decimals
// 	return total_cost, nil
// }