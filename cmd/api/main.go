package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alexalbu001/iguanas-jewelry/internal/auth"
	"github.com/alexalbu001/iguanas-jewelry/internal/handlers"
	"github.com/alexalbu001/iguanas-jewelry/internal/middleware"
	"github.com/alexalbu001/iguanas-jewelry/internal/routes"
	"github.com/alexalbu001/iguanas-jewelry/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
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

	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal("Unable to connect to redis", err)
	}
	rdb := redis.NewClient(opt)

	//create store
	productStore := store.NewProductStore(dbpool)

	userStore := store.NewUsersStore(dbpool)

	cartsStore := store.NewCartsStore(dbpool)

	sessionsStore := auth.NewSessionStore(rdb)

	//create handlers with store
	productHandlers := handlers.NewProductHandlers(productStore)
	userHandlers := handlers.NewUserHandler(userStore)
	authHandlers := auth.NewAuthHandlers(userStore, sessionsStore)
	cartHandlers := handlers.NewCartsHandler(cartsStore, productStore)

	authMiddleware := middleware.NewAuthMiddleware(sessionsStore)
	adminMiddleware := middleware.NewAdminMiddleware(sessionsStore, userStore)

	routes.SetupRoutes(r, productHandlers, userHandlers, cartHandlers, authHandlers, authMiddleware, adminMiddleware)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello World",
		})
	})

	r.Run(os.Getenv("PORT"))

}
