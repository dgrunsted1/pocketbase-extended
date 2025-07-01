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
	IngrNum     string `db:"ingr_num" json:"ingr_num"`
	DirectionsNum     string `db:"directions_num" json:"directions_num"`
}

type Cats struct {
		Name       	string `db:"category" json:"name"`
}

type Cuis struct {
		Name       	string `db:"cuisine" json:"name"`
}


func HandleRandomRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var randomRecipeData RandomRecipeRequest
		var rand_recipe = Rand_Recipe{}
		var cats []Cats
		var cuis []Cuis

		if err := e.BindBody(&randomRecipeData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}



		
			

		if randomRecipeData.Cat != "any" && randomRecipeData.Cuisine != "any"{
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author", "(SELECT COUNT(*) FROM ingredients WHERE id IN (SELECT json_each.value FROM json_each(recipes.ingr_list))) as ingr_num", "json_array_length(directions) as directions_num").
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
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author", "(SELECT COUNT(*) FROM ingredients WHERE id IN (SELECT json_each.value FROM json_each(recipes.ingr_list))) as ingr_num", "json_array_length(directions) as directions_num").
				From("recipes").
				AndWhere(dbx.NewExp("category = {:cat}", dbx.Params{ "cat": randomRecipeData.Cat })).
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		} else if randomRecipeData.Cuisine != "any" {
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author", "(SELECT COUNT(*) FROM ingredients WHERE id IN (SELECT json_each.value FROM json_each(recipes.ingr_list))) as ingr_num", "json_array_length(directions) as directions_num").
				From("recipes").
				AndWhere(dbx.NewExp("cuisine = {:cuisine}", dbx.Params{ "cuisine": randomRecipeData.Cuisine })).
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		} else {
			err := app.DB().
				Select("id", "title", "description", "image", "category", "cuisine", "time", "author", "(SELECT COUNT(*) FROM ingredients WHERE id IN (SELECT json_each.value FROM json_each(recipes.ingr_list))) as ingr_num", "json_array_length(directions) as directions_num").
				From("recipes").
				OrderBy("Random()").
				One(&rand_recipe)

			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
			}
		}

		if randomRecipeData.Cat != "any" {
			err := app.DB().
				Select("cuisine").
				Distinct(true).
				From("recipes").
				AndWhere(dbx.NewExp("category = {:cat}", dbx.Params{ "cat": randomRecipeData.Cat })).
				AndWhere(dbx.NewExp("cuisine <> ''")).
				All(&cuis)

			if err != nil {
				log.Printf("Failed to find cuisines: %v", err)
			}
		} else {
			err := app.DB().
				Select("cuisine").
				Distinct(true).
				From("recipes").
				AndWhere(dbx.NewExp("cuisine <> ''")).
				All(&cuis)

			if err != nil {
				log.Printf("Failed to find cuisines: %v", err)
			}
		}
		
		if randomRecipeData.Cuisine != "any" {
			err := app.DB().
				Select("category").
				Distinct(true).
				From("recipes").
				AndWhere(dbx.NewExp("cuisine = {:cuisine}", dbx.Params{ "cuisine": randomRecipeData.Cuisine })).
				AndWhere(dbx.NewExp("category <> ''")).
				All(&cats)

			if err != nil {
				log.Printf("Failed to find categories: %v", err)
			}
		} else {
			err := app.DB().
				Select("category").
				Distinct(true).
				From("recipes").
				AndWhere(dbx.NewExp("category <> ''")).
				All(&cats)

			if err != nil {
				log.Printf("Failed to find categories: %v", err)
			}
		}

		return e.JSON(http.StatusOK, map[string]interface{}{
            "success":	true,
            "recipe":	rand_recipe,
			"categories": cats,
			"cuisines": cuis,
        })
	}
}