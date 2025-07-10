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
	Note string `json:"note"`
	Author string `json:"author"`
	Category string `json:"category"`
	Country string `json:"country"`
	Cuisine string `json:"cuisine"`
	Description string `json:"description"`
	Directions []string `json:"directions"`
	Expand Expand `json:"expand"`
	Favorite bool `json:"favorite"`
	Image string `json:"image"`
	Servings int `json:"servings"`
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

func HandleSaveRecipe(app *pocketbase.PocketBase) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		var saveData SaveRequest
		
		if err := e.BindBody(&saveData); err != nil {
			return e.BadRequestError("Failed to read request data", err)
		}
		log.Print(saveData)
		err := app.RunInTransaction(func(txApp core.App) error {

			collection, err := txApp.FindCollectionByNameOrId("recipes")
			if err != nil {
				log.Printf("Failed to find collection: %v", err)
				return err
			}
			new_recipe := core.NewRecord(collection)

			mins, err := ParseTimeToMinutes(saveData.Time);
			if (err != nil){
				log.Printf("Failed to parse minutes: %v", err)
				return err
			}

			new_recipe.Set("title", saveData.Title)
			new_recipe.Set("description", saveData.Description)
			new_recipe.Set("user", saveData.UserID)
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
			new_recipe.Set("favorite", saveData.Favorite)
			new_recipe.Set("time_new", mins)
			new_recipe.Set("servings_new", saveData.Servings)
			new_recipe.Set("ingr_num", len(saveData.Expand.IngrList))
			err = txApp.Save(new_recipe);
			if err != nil {
				return err
			}

			ingr_collection, err := txApp.FindCollectionByNameOrId("ingredients")
			if err != nil {
				log.Printf("Failed to find ingredient collection: %v", err)
				return err
			}
			var ingr_ids []string
			for _, curr := range saveData.Expand.IngrList {
				new_ingr := core.NewRecord(ingr_collection)
				new_ingr.Set("quantity", curr.Quantity)
				new_ingr.Set("ingredient", curr.Ingredient)
				new_ingr.Set("unit", curr.Unit)
				new_ingr.Set("unitPlural", curr.UnitPlural)
				new_ingr.Set("symbol", curr.Symbol)
				new_ingr.Set("recipe", new_recipe.Id)
				err = txApp.Save(new_ingr)
				if err != nil {
					return err
				}
				ingr_ids = append(ingr_ids, new_ingr.Id)
			}

			new_recipe.Set("ingr_list", ingr_ids)
			err = txApp.Save(new_recipe)
			if err != nil {
				return err
			}

			return nil
		})
		
		if err != nil {
			log.Printf("Database error: %v", err)
			return e.BadRequestError("Failed to save recipe", err)
		}
		
		return e.JSON(http.StatusOK, map[string]interface{}{
			"status": 200,
			"success": true,
		})
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
		h, err := strconv.ParseFloat(hourMatches[1], 64)
		if err != nil {
			return 0, err
		}
		hours = int(h)
		minutes = int((h-float64(hours))*60)
	}
	
	// Extract minutes if present
	minuteMatches := minuteRegex.FindStringSubmatch(timeStr)
	if len(minuteMatches) > 1 {
		m, err := strconv.ParseFloat(minuteMatches[1], 64)
		if err != nil {
			return 0, err
		}
		minutes += int(m)
	}
	
	// If neither hours nor minutes were found, try parsing just a number as minutes
	if hours == 0 && minutes == 0 {
		// Try to extract just a number (assuming it's minutes)
		numRegex := regexp.MustCompile(`^\s*(\d+)\s*$`)
		numMatches := numRegex.FindStringSubmatch(timeStr)
		if len(numMatches) > 1 {
			m, err := strconv.ParseFloat(numMatches[1], 64)
			if err != nil {
				return 0, err
			}
			minutes = int(m)
		}
	}
	
	// Calculate total minutes
	totalMinutes := (hours * 60) + minutes
	
	return totalMinutes, nil
}