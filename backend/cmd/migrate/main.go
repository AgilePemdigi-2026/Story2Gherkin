
package main

import (
    "log"

    "github.com/story2gherkin/config"
    "github.com/story2gherkin/database"
)

func main() {
    db := config.SetUpDatabaseConnection()

    if err := database.Migrate(db); err != nil {
        log.Fatalf("Migration failed: %v", err)
    }

    log.Println("Migration completed successfully")
}
