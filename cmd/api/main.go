package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/routes"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func init() {

}

func main() {
	r := gin.Default()

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to db", err)
	}
	// Verify the connection
	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatal("Unable to ping database:", err)
	}

	fmt.Println("Connected to PostgreSQL database!")

	defer dbpool.Close()
	//create store
	productStore := store.NewProductStore(dbpool)

	//create handlers with store
	productHandlers := &handlers.ProductHandlers{
		Store: productStore,
	}

	routes.SetupRoutes(r, productHandlers)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run(os.Getenv("PORT"))

}
