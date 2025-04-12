package handlers

import (
    "log"
    "net/http"
    "strings"
    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    
)

type AddRequest struct {
    ID          string   `json:"id"`
}

// type Recipe struct {
// 	Title       string `db:"title" json:"title"`
// 	Id       string `db:"id" json:"id"`
// 	Author 		string `db:"author" json:"author"`
// 	Time        string `db:"time" json:"time"`
// 	Image       string `db:"image" json:"image"`
// 	UrlId       string `db:"url_id" json:"url_id"`
// 	Category    string `db:"category" json:"category"`
// 	Cuisine     string `db:"cuisine" json:"cuisine"`
// 	Country     string `db:"country" json:"country"`
// 	Directions 	string `db:"directions" json:"directions"`
// 	IngrList 	string `db:"ingr_list" json:"ingr_list"`
// 	Servings 	string `db:"servings" json:"servings"`
// 	User 	string `db:"user" json:"user"`
// }

// type Category struct {
// 	Id       string `db:"id" json:"id"`
// }

func get_copy_query(addData AddRequest) string {
	var query = `
		BEGIN TRANSACTION;

		-- First, insert a copy of the recipe with the new user
		WITH new_recipe AS (
		INSERT INTO recipes (
			-- All columns except primary key and the ingredients relation
			name, description, instructions, cooking_time, user_id, -- other fields...
		)
		SELECT 
			name, description, instructions, cooking_time, 
			NEW_USER_ID, -- Replace with your new user's ID
			-- other fields...
		FROM recipes 
		WHERE id = ORIGINAL_RECIPE_ID
		RETURNING id -- Return the new recipe ID
		)

		-- Then, copy the ingredients associated with the original recipe
		INSERT INTO ingredients (
		-- All columns except primary key and the recipe relation
		name, quantity, unit, recipe_id, -- other fields...
		)
		SELECT 
		i.name, i.quantity, i.unit, 
		(SELECT id FROM new_recipe), -- Use the newly created recipe ID
		-- other fields...
		FROM ingredients i
		WHERE i.recipe_id = ORIGINAL_RECIPE_ID;

		COMMIT;`

	return query;
}

func HandleCopy(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
    return func(e *core.RequestEvent) error {
        var addData AddRequest

        if err := e.BindBody(&addData); err != nil {
            return e.BadRequestError("Failed to read request data", err)
        }
        
        var params = dbx.Params{
            "id": addData.ID,
		}

        var recipeQuery = get_copy_query(addData)
        recipes := []Recipe{}
		app.DB().
        if err := app.DB().NewQuery(recipeQuery).Bind(params).All(&recipes); err != nil {
            log.Printf("Database error: %v", err)
        }

        

        return e.JSON(http.StatusOK, map[string]interface{}{
            "success":	true,
        })
    }
}