// main.go
package main

import (
    "log"
    
    "github.com/pocketbase/pocketbase"
    
    "pocketbase-extended/routes" // Adjust this import path
)

func main() {
    app := pocketbase.New()
    
    // Register all routes
    routes.RegisterRoutes(app)
    
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}