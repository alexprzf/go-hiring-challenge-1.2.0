package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
)

type CombinedRepository struct {
	*models.ProductsRepository
	*models.CategoriesRepository
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, close := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	defer close()

	// Initialize handlers
	prodRepo := models.NewProductsRepository(db)
	categoriesRepo := models.NewCategoriesRepository(db)

	repo := &CombinedRepository{
		ProductsRepository:   prodRepo,
		CategoriesRepository: categoriesRepo,
	}

	cat := catalog.NewCatalogHandler(repo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", cat.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", cat.HandleGetDetails)
	mux.HandleFunc("GET /categories", cat.HandleGetCategories)
	mux.HandleFunc("POST /categories", cat.HandleCreateCategory)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", os.Getenv("HTTP_PORT")),
		Handler: mux,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://%s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Shutdown(ctx)
	stop()
}
