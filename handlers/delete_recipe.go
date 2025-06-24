package handlers

import (
	"net/http"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"log"
	"github.com/pocketbase/dbx"
)

type DeleteRecipeRequest struct {
	RecipeId string `json:"recipe_id"`
}

func HandleDeleteRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var deleteRecipeData DeleteRecipeRequest
		
		if err := e.BindBody(&deleteRecipeData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		//get refferences to the recipe
		recipe, err := app.FindRecordById("recipes", deleteRecipeData.RecipeId)
		if err != nil {
			log.Printf("Failed to find recipe: %v", err)
			return err
		}

		logs, err := app.FindAllRecords("recipe_log", dbx.NewExp("recipe = {:recipe_id}", dbx.Params{"recipe_id": deleteRecipeData.RecipeId}))
		if err != nil {
			log.Printf("Failed to find logs: %v", err)
			return err
		}

		menus, err := app.FindAllRecords("menus", dbx.Like("recipes", deleteRecipeData.RecipeId))
		if err != nil {
			log.Printf("Failed to find menus: %v", err)
			return err
		}
		app.RunInTransaction(func(txApp core.App) error {
			for _, curr := range recipe.GetStringSlice("ingr_list") {
				ingr, err := app.FindRecordById("ingredients", curr)
				if err != nil {
					return err
				}
				err = txApp.Delete(ingr)
				if err != nil {
					return err
				}
			}

			for _, curr := range logs {
				err = txApp.Delete(curr)
				if err != nil {
					return err
				}
			}

			for _, curr := range menus {
				curr.Set("recipes-", curr.Id)
			}

			txApp.Delete(recipe)
			return nil
		})

		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
		})
	}
}