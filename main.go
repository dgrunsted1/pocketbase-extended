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

type Category struct {
	Id       string `db:"id" json:"id"`
}

func get_recipe_query(searchData SearchRequest) string {
	var query = `SELECT r.id, r.title, r.author, r.time, r.image, r.category, r.url_id, r.cuisine, r.country, (select count(*) from json_each(r.directions)) as directions, (select count(*) from json_each(r.ingr_list)) as ingr_list, r.servings, r.user
					`+search_tables()+`
					WHERE made = 1`+search_filters("", searchData)

	query += ` group by r.id`

	switch searchData.SortVal {
		case "Least Ingredients":
			query += ` order by ingr_list asc`
		case "Most Ingredients":
			query += ` order by ingr_list desc`
		case "Least Servings":
			query += ` order by r.servings asc`
		case "Most Servings":
			query += ` order by r.servings desc`
		case "Least Time":
			query += ` order by r.time_new asc`
		case "Most Time":
			query += ` order by r.time_new desc`
		case "Most Recent":
			query += ` order by r.created asc`
		case "Least Recent":
			query += ` order by r.created desc`
		default:
			query += ` order by r.created asc`
	}
	return query;
}

func search_tables() string {
	return ` FROM recipes r 
			JOIN json_each(r.ingr_list) AS je 
			JOIN ingredients i ON i.id = je.value`
}

func search_filters(cat_type string, searchData SearchRequest) string {
	print(cat_type)
	var query = ` and made = 1`
	if (len(searchData.SelectedCategories) > 0 && cat_type != "category") {
		query += ` and r.category IN (`
		for i := range searchData.SelectedCategories {
			if i == len(searchData.SelectedCategories) - 1 {
				query += `'` + searchData.SelectedCategories[i] + `'`
			} else {
				query += `'` + searchData.SelectedCategories[i] + `',`
			}
		}
		query += `)`
	}
	if len(searchData.SelectedCuisines) > 0 && cat_type != "cuisine" {
		query += ` and r.cuisine IN (`
		for i := range searchData.SelectedCuisines {
			if i == len(searchData.SelectedCuisines) - 1 {
				query += `'` + searchData.SelectedCuisines[i] + `'`
			} else {
				query += `'` + searchData.SelectedCuisines[i] + `',`
			}
		}
		query += `)`
	}
	if len(searchData.SelectedCountries) > 0 && cat_type != "country" {
		query += ` and r.country IN (`
		for i := range searchData.SelectedCountries {
			if i == len(searchData.SelectedCountries) - 1 {
				query += `'` + searchData.SelectedCountries[i] + `'`
			} else {
				query += `'` + searchData.SelectedCountries[i] + `',`
			}
		}
		query +=`)`
	}
	if len(searchData.SelectedAuthors) > 0 && cat_type != "authors" {
		query += ` and r.author IN (`
		for i := range searchData.SelectedAuthors {
			if i == len(searchData.SelectedAuthors) - 1 {
				query += `'` + searchData.SelectedAuthors[i] + `'`
			} else {
				query += `'` + searchData.SelectedAuthors[i] + `',`
			}
		}
		query += `)`
	}
	if searchData.SortVal == "Least Time" || searchData.SortVal == "Most Time" {
		query += ` and r.time_new <> "" and r.time_new <> 0 and r.time <> ""`
	}
	if len(searchData.SearchVal) > 0 {query += ` and (i.ingredient LIKE '%' || {:search_val} || '%' OR r.title LIKE '%' || {:search_val} || '%')`}
	return query
}

func get_categories_query(cat_type string, searchData SearchRequest) string {
	var query = `SELECT Distinct r.`+cat_type+` as id`+search_tables()+`
					WHERE r.`+cat_type+` <> ""`+search_filters(cat_type, searchData)

	query += ` order by r.`+cat_type+` asc`
	return query;
}



func main() {
	app := pocketbase.New()
	
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		se.Router.POST("/api/search", func(e *core.RequestEvent) error {
			var searchData SearchRequest

			if err := e.BindBody(&searchData); err != nil {
				return e.BadRequestError("Failed to read request data", err)
			}
			var params = dbx.Params{
				"search_val": searchData.SearchVal,
				"selected_categories": strings.Join(searchData.SelectedCategories, ","),
				"selected_cuisines": strings.Join(searchData.SelectedCuisines, ","),
				"selected_countries": strings.Join(searchData.SelectedCountries, ","),
				"selected_authors": strings.Join(searchData.SelectedAuthors, ","),
			}

			var recipe_query = get_recipe_query(searchData)
			recipes := []Recipe{}
			if err := app.DB().NewQuery(recipe_query).Bind(params).All(&recipes); err != nil {
				log.Printf("Database error: %v", err)
			}

			var categories_query = get_categories_query(`category`, searchData)
			categories := []Category{}
			if err := app.DB().NewQuery(categories_query).Bind(params).All(&categories); err != nil {
				log.Printf("Database error: %v", err)
			}

			var countries_query = get_categories_query(`country`, searchData)
			countries := []Category{}
			if err := app.DB().NewQuery(countries_query).Bind(params).All(&countries); err != nil {
				log.Printf("Database error: %v", err)
			}

			var cuisines_query = get_categories_query(`cuisine`, searchData)
			cuisines := []Category{}
			if err := app.DB().NewQuery(cuisines_query).Bind(params).All(&cuisines); err != nil {
				log.Printf("Database error: %v", err)
			}

			var authors_query = get_categories_query(`author`, searchData)
			authors := []Category{}
			if err := app.DB().NewQuery(authors_query).Bind(params).All(&authors); err != nil {
				log.Printf("Database error: %v", err)
			}

            return e.JSON(http.StatusOK, map[string]interface{}{
                "success":             true,
				"recipes":             recipes,
				"categories":          categories,
				"countries":           countries,
				"cuisines":           	cuisines,
				"authors":             authors,
				"page": 			1,
				"perPage": 			len(recipes),
				"totalItems":		len(recipes),
				"totalPages":		1,
            })
		})
	
		return se.Next()
	})
	
	
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}