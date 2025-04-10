package main

import (
	"log"
	"net/http"
	"strings"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

type SearchRequest struct {
    SearchVal          string   `json:"search_val"`
	SortVal            string   `json:"sort_val"`
	SelectedCategories []string `json:"selected_categories"`
	SelectedCountries  []string `json:"selected_countries"`
	SelectedCuisines   []string `json:"selected_cuisines"`
	SelectedAuthors    []string `json:"selected_authors"`
	Page               int      `json:"page"`
	PerPage            int      `json:"per_page"`
}

type Recipe struct {
	Title       string `db:"title" json:"title"`
	Id       string `db:"id" json:"id"`
	Author 		string `db:"author" json:"author"`
	Time        string `db:"time" json:"time"`
	Image       string `db:"image" json:"image"`
	UrlId       string `db:"url_id" json:"url_id"`
	Category    string `db:"category" json:"category"`
	Cuisine     string `db:"cuisine" json:"cuisine"`
	Country     string `db:"country" json:"country"`
	Directions 	string `db:"directions" json:"directions"`
	IngrList 	string `db:"ingr_list" json:"ingr_list"`
	Servings 	string `db:"servings" json:"servings"`
	User 	string `db:"user" json:"user"`
}

func main() {
	app := pocketbase.New()
	
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/search", func(e *core.RequestEvent) error {
			var searchData SearchRequest
			
            if err := e.BindBody(&searchData); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}

			recipes := []Recipe{}
			var query = `SELECT r.id, r.title, r.author, r.time, r.image, r.category, r.url_id, r.cuisine, r.country, (select count(*) from json_each(r.directions)) as directions, (select count(*) from json_each(r.ingr_list)) as ingr_list, r.servings, r.user
							FROM recipes r
							JOIN json_each(r.ingr_list) AS je
							JOIN ingredients i ON i.id = je.value
							WHERE made = 1`
			if len(searchData.SearchVal) > 0 {query += ` and (i.ingredient LIKE '%' || {:search_val} || '%' OR r.title LIKE '%' || {:search_val} || '%')`}
			if len(searchData.SelectedCategories) > 0 {
				query += ` and r.category IN ({:selected_categories})`
			}
			if len(searchData.SelectedCuisines) > 0 {
				query += ` and r.cuisine IN ({:selected_cuisines})`
			}
			if len(searchData.SelectedCountries) > 0 {
				query += ` and r.country IN ({:selected_countries})`
			}
			if len(searchData.SelectedAuthors) > 0 {
				query += ` and r.author IN ({:selected_authors})`
			}
			query += ` group by r.id`
			err := app.DB().
					NewQuery(query).
					Bind(dbx.Params{
						"search_val": searchData.SearchVal,
						"selected_categories": strings.Join(searchData.SelectedCategories, ","),
						"selected_cuisines": strings.Join(searchData.SelectedCuisines, ","),
						"selected_countries": strings.Join(searchData.SelectedCountries, ","),
						"selected_authors": strings.Join(searchData.SelectedAuthors, ","),
					}).All(&recipes)

			

			if err != nil {
				log.Printf("Database error: %v", err)
  				return e.BadRequestError("Database query failed", err)
			}
			print(len(recipes))
            return e.JSON(http.StatusOK, map[string]interface{}{
                "success":             true,
				"count":               len(recipes),
				"recipes":             recipes,
            })
		})
	
		return se.Next()
	})
	
	
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}