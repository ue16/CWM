package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "oskar/cwm25/docs"
	"oskar/cwm25/internal"
)

//	@title			CWM Kanban
//	@version		0.3
//	@description	This is a small kanban server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath	/
func main() {
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL env var not set")
	}

	db, err := internal.NewDB(connString)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Println("SECRET env var not set; falling back to default")
		secret = "super-secure-hardcoded-secret-for-testing-purposes-only"
	}
	addr := fmt.Sprintf(":%s", port)
	handler := internal.BasicServer(addr, secret, db).Docs().Finalize()

	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
