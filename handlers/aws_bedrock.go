package handlers

import (
	"net/http"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type GroceryListRequest struct {
	List string `json:"list"`
}

func HandleGroceryList(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var groceryListData GroceryListRequest

		if err := e.BindBody(&groceryListData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}
		// add bedrock request handling logic here
		
		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
			"data": groceryListData,
		})
	}
}