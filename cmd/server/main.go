package main

import (
	"internal-transfers/internal/config"
	"internal-transfers/internal/handlers"
	"internal-transfers/internal/repository"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.LoadConfig()

	db, err := config.ConnectDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	r := mux.NewRouter()

	r.HandleFunc("/accounts", handler.CreateAccount).Methods("POST")
	r.HandleFunc("/accounts/{accountID}", handler.GetAccountBalance).Methods("GET")
	r.HandleFunc("/transactions", handler.SubmitTransaction).Methods("POST")
	r.HandleFunc("/transfers", handler.CreateTransfer).Methods("POST")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}