package main 

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