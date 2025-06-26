package handlers

import (
	"net/http"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type RandomRecipeRequest struct {
	Cat string `json:"cat"`
	Cuisine string `json:"cuisine"`
}

func HandleRandomRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var randomRecipeData RandomRecipeRequest
		
		if err := e.BindBody(&randomRecipeData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		

		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
			"data": randomRecipeData,
		})
	}
}