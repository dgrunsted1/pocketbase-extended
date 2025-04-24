package handlers

import (
	"log"
	"net/http"
	// "github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"regexp"
	"strconv"
	"strings"
)

type SaveRequest struct {
	UserID   string `json:"user_id"`
	User   string `json:"user"`
	Note string `json:"note"`
	Author string `json:"author"`
	Category string `json:"category"`
	Country string `json:"country"`
	Cuisine string `json:"cuisine"`
	Description string `json:"description"`
	Directions []string `json:"directions"`
	Expand Expand `json:"expand"`
	Favorite bool `json:"favorite"`
	Id string `json:"id"`
	Image string `json:"image"`
	Made bool `json:"made"`
	Servings string `json:"servings"`
	Time string `json:"time"`
	Title string `json:"title"`
	Url string `json:"url"`
}

type Expand struct {
    IngrList []Ingredient `json:"ingr_list"`
}

type Ingredient struct {
    CollectionId   string   `json:"collectionId"`
    CollectionName string   `json:"collectionName"`
    Created        string   `json:"created"`
    Id             string   `json:"id"`
    Ingredient     string   `json:"ingredient"`
    Quantity       float64  `json:"quantity"`
    Recipe         []string `json:"recipe,omitempty"`
    Symbol         string   `json:"symbol"`
    Unit           string   `json:"unit"`
    UnitPlural     string   `json:"unitPlural"`
    Updated        string   `json:"updated"`
}
// type AddResult struct {
// 	Id string `db:"id" json:"id"`
// }

// type AddRecipe struct {
// 	Title       string `db:"title" json:"title"`
// 	Id       	string `db:"id" json:"id"`
// 	Description	string `db:"description" json:"description"`
// 	Url			string `db:"url" json:"url"`
// 	Author 		string `db:"author" json:"author"`
// 	Time        string `db:"time" json:"time"`
// 	Image       string `db:"image" json:"image"`
// 	Category    string `db:"category" json:"category"`
// 	Cuisine     string `db:"cuisine" json:"cuisine"`
// 	Country     string `db:"country" json:"country"`
// 	Directions 	string `db:"directions" json:"directions"`
// 	IngrList 	string `db:"ingr_list" json:"ingr_list"`
// 	Servings 	string `db:"servings" json:"servings"`
// 	User 		string `db:"user" json:"user"`
// 	Notes 		string `db:"notes" json:"notes"`
// 	TimeNew 	string `db:"time_new" json:"time_new"`
// 	IngrNum 	string `db:"ingr_num" json:"ingr_num"`
// 	ServingsNew	string `db:"servings_new" json:"servings_new"`
// }

func HandleSaveRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var saveData SaveRequest
		
		if err := e.BindBody(&saveData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}

		if (saveData.UserID != saveData.User) {
			return e.BadRequestError("Failed to save recipe", "unauthorized")
		}

		err := app.RunInTransaction(func(txApp core.App) error {

			// collection, err := txApp.FindCollectionByNameOrId("recipes")
			// if err != nil {
			// 	log.Printf("Failed to find collection: %v", err)
			// 	return err
			// }
			new_recipe, err := txApp.FindRecordById("recipes", saveData.Id)
			if err != nil {
				log.Printf("Failed to find recipe: %v", err)
				return err
			}

			mins, err := ParseTimeToMinutes(saveData.Time);
			if (err != nil){
				log.Printf("Failed to parse minutes: %v", err)
				return err
			}
			servings, err := get_servings_int(saveData.Servings)
			if (err != nil) {
				log.Printf("Failed to parse servings: %v", err)
				return err
			}

			new_recipe.Set("title", saveData.Title)
			new_recipe.Set("description", saveData.Description)
			new_recipe.Set("url", saveData.Url)
			new_recipe.Set("author", saveData.Author)
			new_recipe.Set("time", saveData.Time)
			new_recipe.Set("directions", saveData.Directions)
			new_recipe.Set("image", saveData.Image)
			new_recipe.Set("servings", saveData.Servings)
			new_recipe.Set("cuisine", saveData.Cuisine)
			new_recipe.Set("country", saveData.Country)
			new_recipe.Set("notes", saveData.Note)
			new_recipe.Set("category", saveData.Category)
			new_recipe.Set("made", saveData.Made)
			new_recipe.Set("favorite", saveData.Favorite)
			new_recipe.Set("time_new", mins)
			new_recipe.Set("servings_new", servings)
			new_recipe.Set("ingr_num", len(saveData.Expand.IngrList))
			err = txApp.Save(new_recipe);
			if err != nil {
				return err
			}
		// 	// Step 2: Copy the ingredients associated with the original recipe to the new recipe
		// 	insertIngredientsSQL := `
		// 		INSERT INTO ingredients (
		// 			quantity, ingredient, unit, unitPlural, symbol, recipe
		// 		)
		// 		SELECT
		// 			i.quantity, i.ingredient, i.unit, i.unitPlural, i.symbol, {:new_recipe_id}
		// 		FROM recipes r 
		// 		JOIN json_each(r.ingr_list) AS je 
		// 		JOIN ingredients i ON i.id = je.value
		// 		WHERE r.id = {:original_recipe_id}
		// 	`
		// 	var params = dbx.Params{
		// 		"original_recipe_id":       original_recipe.Id,
		// 		"new_recipe_id":            new_recipe.Id,
		// 	};

		// 	var result = []AddResult{}
		// 	if err := txApp.DB().NewQuery(insertIngredientsSQL).Bind(params).All(&result); err != nil {
		// 		return err
		// 	}
			
		// 	// Step 3: Update the ingr_list in the new recipe to include the newly inserted ingredients
		// 	ingredientIds, err := txApp.FindAllRecords("ingredients", dbx.HashExp{"recipe": new_recipe.Id})
		// 	if err != nil {
		// 		return err
		// 	}

		// 	for _, curr := range ingredientIds {
		// 		new_recipe.Set("ingr_list+", curr.Id)
		// 	}
		// 	err = txApp.Save(new_recipe);
		// 	if err != nil {
		// 		return err
		// 	}
			return e.JSON(http.StatusOK, map[string]interface{}{})
		})
		
		if err != nil {
			log.Printf("Database error: %v", err)
			return e.BadRequestError("Failed to save recipe", err)
		}
		
		return e.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"data": saveData,
		})
	}
}

func get_servings_int(serving_str string) (int, error) {
	serving_str = strings.ToLower(serving_str)

	servings_regex := regexp.MustCompile(`(\d+(?:\.\d+)?)`)

	servings_matches := servings_regex.FindStringSubmatch(serving_str)

	if len(servings_matches) > 1 {
		h, err := strconv.Atoi(servings_matches[1])
		if err != nil {
			return 0, err
		}
		return int(h), nil
	}else {
		return 0, nil
	}
}

// ParseTimeToMinutes takes a time string like "20 minutes" or "1 hr 12 mins"
// and returns the total number of minutes as an integer
func ParseTimeToMinutes(timeStr string) (int, error) {
	// Convert to lowercase for easier matching
	timeStr = strings.ToLower(timeStr)
	
	timeStr = strings.ReplaceAll(timeStr, "½", ".5")
	timeStr = strings.ReplaceAll(timeStr, "¼", ".25")
	timeStr = strings.ReplaceAll(timeStr, "¾", ".75")
	timeStr = strings.ReplaceAll(timeStr, "⅓", ".33")
	timeStr = strings.ReplaceAll(timeStr, "⅔", ".67")

	// Initialize variables for hours and minutes
	hours := 0
	minutes := 0
	
	// Regular expressions for matching hours and minutes
	hourRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:hr|hour|hours|h)`)
	minuteRegex := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:min|mins|minutes|m)`)
	
	// Extract hours if present
	hourMatches := hourRegex.FindStringSubmatch(timeStr)
	if len(hourMatches) > 1 {
		h, err := strconv.Atoi(hourMatches[1])
		if err != nil {
			return 0, err
		}
		hours = h
	}
	
	// Extract minutes if present
	minuteMatches := minuteRegex.FindStringSubmatch(timeStr)
	if len(minuteMatches) > 1 {
		m, err := strconv.Atoi(minuteMatches[1])
		if err != nil {
			return 0, err
		}
		minutes = m
	}
	
	// If neither hours nor minutes were found, try parsing just a number as minutes
	if hours == 0 && minutes == 0 {
		// Try to extract just a number (assuming it's minutes)
		numRegex := regexp.MustCompile(`^\s*(\d+)\s*$`)
		numMatches := numRegex.FindStringSubmatch(timeStr)
		if len(numMatches) > 1 {
			m, err := strconv.Atoi(numMatches[1])
			if err != nil {
				return 0, err
			}
			minutes = m
		}
	}
	
	// Calculate total minutes
	totalMinutes := (hours * 60) + minutes
	
	return totalMinutes, nil
}