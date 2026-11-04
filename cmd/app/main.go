package main
import (
	"fmt"
	"log"
	"net/http"

	"social_network/internal/db"
	"social_network/internal/handlers"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatal(err)
	}
	defer db.Pool.Close()

	h := handlers.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /user/register", h.Register)
	mux.HandleFunc("POST /login", h.Login)
	// GetUser ...  middleware
	mux.HandleFunc("GET /user/get/{id}", h.AuthMiddleware(h.GetUser))
	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}