package model

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// processDbEnvVar correctly escapes values for the Postgres connection string specified in
// https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
func processDbEnvVar(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		log.Printf("Variable `%s` not set.", varName)
	}
	// The connection string requires that single quotes are escaped
	return strings.ReplaceAll(value, "'", "\\'")
}

// buildConnStr
func buildConnStr() string {
	dbName := processDbEnvVar("DATABASE_NAME")
	dbUser := processDbEnvVar("DATABASE_USER")
	dbPassword := processDbEnvVar("DATABASE_PASSWORD")
	dbHost := processDbEnvVar("DATABASE_HOST")

	return fmt.Sprintf("dbname='%s' user='%s' password='%s' host='%s' connect_timeout=1 sslmode=disable", dbName, dbUser, dbPassword, dbHost)
}

// twoToPow raises 2 to the power i - i.e. 2**i
func twoToPow(i int) int {
	return int(1 << uint(i))
}

// connectionBackoff tries to run a query on the database ad if it fails, tries again after an increasing wait.
//
// This is useful because when deploying, it is often the case that the database only becomes accessible some time
// after deployment. (e.g. because the CloudSQL proxy takes time to establish the connection to the database.)
func connectionBackoff(db *sql.DB, maxAttempts int) error {
	attempt := 1

	_, err := db.Exec("SET timezone = 'utc'")
	for err != nil && attempt < maxAttempts {
		log.Printf("Cannot connect to DB, backing off and trying again in %d seconds. (%v)", twoToPow(attempt), err)

		// Back off doubling the wait time every time. Start with a wait of 2 seconds (2**1)
		time.Sleep(time.Duration(twoToPow(attempt)) * time.Second)
		attempt++
		_, err = db.Query("SET timezone = 'utc'")
	}
	if attempt >= maxAttempts {
		log.Fatalf("Unable to connect to Database. (%v)", err)
	}
	return err
}

// initDb runs any necesary database initialization. For example creating tables or running migrations.
func initDb(db *sql.DB) error {

	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS resource_metadata (
			id          TEXT NOT NULL,
			type        TEXT NOT NULL,
			created_at  TIMESTAMP NOT NULL,
			updated_at  TIMESTAMP NOT NULL,
			deleted_at  TIMESTAMP,
			params      JSONB NOT NULL,
			data        JSONB NOT NULL,
			PRIMARY KEY (id)
	)`)
	if err != nil {
		log.Println("Unable to create resource_metadata table.")
		return fmt.Errorf("create resource_metadata table: %w", err)
	}

	return nil
}

// Setup attempts to connect to the database and then run any initialization.
func Setup() Modeler {
	log.Println("Connecting to Database.")
	db, err := sql.Open("postgres", buildConnStr())
	if err != nil {
		log.Fatal(err)
	}

	// Block executing while we attempt to connect to the database
	connectionBackoff(db, 6)

	log.Println("Initializing Database.")
	// Run necessary db commands e.g. migrations
	initDb(db)

	return model{db}
}
