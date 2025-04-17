package handlers

import (
	"log"
	"net/http"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type AddRequest struct {
	UserID   string `json:"user_id"`
	RecipeID string `json:"recipe_id"`
}

type AddResult struct {
	Id string `db:"id" json:"id"`
}

type AddRecipe struct {
	Title       string `db:"title" json:"title"`
	Id       	string `db:"id" json:"id"`
	Description	string `db:"description" json:"description"`
	Url			string `db:"url" json:"url"`
	Author 		string `db:"author" json:"author"`
	Time        string `db:"time" json:"time"`
	Image       string `db:"image" json:"image"`
	Category    string `db:"category" json:"category"`
	Cuisine     string `db:"cuisine" json:"cuisine"`
	Country     string `db:"country" json:"country"`
	Directions 	string `db:"directions" json:"directions"`
	IngrList 	string `db:"ingr_list" json:"ingr_list"`
	Servings 	string `db:"servings" json:"servings"`
	User 		string `db:"user" json:"user"`
	Notes 		string `db:"notes" json:"notes"`
	TimeNew 	string `db:"time_new" json:"time_new"`
	IngrNum 	string `db:"ingr_num" json:"ingr_num"`
	ServingsNew	string `db:"servings_new" json:"servings_new"`
}

func HandleCopy(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var addData AddRequest
		if err := e.BindBody(&addData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		// Start a transaction
		err := app.RunInTransaction(func(txApp core.App) error {
			// Step 1: Insert a copy of the recipe for the new user
			collection, err := txApp.FindCollectionByNameOrId("recipes")
			if err != nil {
				log.Printf("Failed to find collection: %v", err)
				return err
			}
			new_recipe := core.NewRecord(collection)
			
			
			original_recipe, err := txApp.FindRecordById("recipes", addData.RecipeID)

			if err != nil {
				log.Printf("Failed to find record: %v", err)
				return err
			}

			new_recipe.Set("title", original_recipe.Get("title"))
			new_recipe.Set("description", original_recipe.Get("description"))
			new_recipe.Set("url", original_recipe.Get("url"))
			new_recipe.Set("author", original_recipe.Get("author"))
			new_recipe.Set("time", original_recipe.Get("time"))
			new_recipe.Set("directions", original_recipe.Get("directions"))
			new_recipe.Set("user", addData.UserID)
			new_recipe.Set("image", original_recipe.Get("image"))
			new_recipe.Set("servings", original_recipe.Get("servings"))
			new_recipe.Set("cuisine", original_recipe.Get("cuisine"))
			new_recipe.Set("country", original_recipe.Get("country"))
			new_recipe.Set("notes", original_recipe.Get("notes"))
			new_recipe.Set("category", original_recipe.Get("category"))
			// new_recipe.Set("url_id", original_recipe.Get("country"))
			new_recipe.Set("made", false)
			new_recipe.Set("favorite", false)
			new_recipe.Set("time_new", original_recipe.Get("time_new"))
			new_recipe.Set("servings_new", original_recipe.Get("servings_new"))
			new_recipe.Set("ingr_num", original_recipe.Get("ingr_num"))
			err = app.Save(new_recipe);
			if err != nil {
				return err
			}
			// Step 2: Copy the ingredients associated with the original recipe to the new recipe
			insertIngredientsSQL := `
				INSERT INTO ingredients (
					quantity, ingredient, unit, unitPlural, symbol, recipe
				)
				SELECT
					i.quantity, i.ingredient, i.unit, i.unitPlural, i.symbol, {:new_recipe_id}
				FROM recipes r 
				JOIN json_each(r.ingr_list) AS je 
				JOIN ingredients i ON i.id = je.value
				WHERE r.id = {:original_recipe_id}
			`
			var params = dbx.Params{
				"original_recipe_id":       original_recipe.Id,
				"new_recipe_id":            new_recipe.Id,
			};

			result := AddResult{}
			if err := txApp.DB().NewQuery(insertIngredientsSQL).Bind(params).All(&result); err != nil {
				return err
			}
			
			// Step 3: Update the ingr_list in the new recipe to include the newly inserted ingredients
			// updateIngrListSQL := `
			// 	UPDATE Recipe
			// 	SET ingr_list = (
			// 		SELECT group_concat(id, ',')
			// 		FROM ingredients
			// 		WHERE recipe = {:new_recipe_id}
			// 	)
			// 	WHERE id = {:new_recipe_id}
			// `
			
			// if err := txApp.DB().NewQuery(updateIngrListSQL).Bind(params).All(&result); err != nil {
			// 	return err
			// }
			return e.JSON(http.StatusOK, map[string]interface{}{
				"recipe": original_recipe,
			})
		})
		
		if err != nil {
			log.Printf("Database error: %v", err)
			return e.BadRequestError("Failed to copy recipe", err)
		}
		
		return e.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
		})
	}
}