package Database

import (
	"context"

	"gitlab.com/resticgui/resticgui/ent"

	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

func Connect() (_db *ent.Client) {
	dbname := viper.GetString("db.name")
	client, err := ent.Open("sqlite3", dbname+"ent?_fk=1")
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	// defer client.Close()
	// Run the auto migration tool.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
	return client
}

func Migrate() {
	db := Connect()
	if err := db.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
