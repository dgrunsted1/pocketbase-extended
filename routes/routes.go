// routes/routes.go
package routes

import (
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    
    "pocketbase-extended/handlers" // Adjust this import path
)

func RegisterRoutes(app *pocketbase.PocketBase) {
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.POST("/api/search", handlers.HandleSearch(app))
        
        return se.Next()
    })

    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.POST("/api/copy", handlers.HandleCopy(app))
        
        return se.Next()
    })

    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.POST("/api/save_recipe", handlers.HandleSaveRecipe(app))
        
        return se.Next()
    })
}