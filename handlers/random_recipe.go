package handlers

import (
	"log"
	"net/http"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type RandomRecipeRequest struct {
	Cat string `json:"cat"`
	Cuisine string `json:"cuisine"`
}

type Rand_Recipe struct {
	Title       string `db:"title" json:"title"`
	Id       	string `db:"id" json:"id"`
	Author 		string `db:"author" json:"author"`
	Description	string `db:"description" json:"description"`
	Time        string `db:"time" json:"time"`
	Image       string `db:"image" json:"image"`
	Category    string `db:"category" json:"category"`
	Cuisine     string `db:"cuisine" json:"cuisine"`
}

func HandleRandomRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var randomRecipeData RandomRecipeRequest
		var rand_recipe = Rand_Recipe{}

		if err := e.BindBody(&randomRecipeData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}



		
			

		if randomRecipeData.Cat != "any" && randomRecipeData.Cuisine != "any"{
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author").
				From("recipes").
				AndWhere(dbx.NewExp("category = {:cat}", dbx.Params{ "cat": randomRecipeData.Cat })).
				AndWhere(dbx.NewExp("cuisine = {:cuisine}", dbx.Params{ "cuisine": randomRecipeData.Cuisine })).
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		} else if randomRecipeData.Cat != "any" {
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author").
				From("recipes").
				AndWhere(dbx.NewExp("LOWER(category) = {:cat}", dbx.Params{ "cat": randomRecipeData.Cat })).
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		} else if randomRecipeData.Cuisine != "any" {
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author").
				From("recipes").
				AndWhere(dbx.NewExp("cuisine = {:cuisine}", dbx.Params{ "cuisine": randomRecipeData.Cuisine })).
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		}
		return e.JSON(http.StatusOK, map[string]interface{}{
            "success":	true,
            "recipe":	rand_recipe,
        })
	}
}