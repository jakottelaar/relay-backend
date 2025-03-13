package infra

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/jakottelaar/relay-backend/internal/relationships"
	"github.com/jakottelaar/relay-backend/internal/users"
)

type App struct {
	HttpServer *http.Server
	config     *config.Config
	db         *sql.DB
}

func NewApp(ctx context.Context, config *config.Config) (*App, error) {
	db, err := initializeDB(config.DSN)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	log.Println("database connection established")

	router := gin.Default()
	router.Use(
		gin.Recovery(),
	)

	registerRoutes(router, db, *config)

	log.Println("routes registered")

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		HttpServer: srv,
		config:     config,
		db:         db,
	}, nil
}

func registerRoutes(r *gin.Engine, db *sql.DB, cfg config.Config) {

	r.Use(internal.ErrorHandler())

	r.GET("/health", handleHealth(db))

	userRepo := users.NewUserRepo(db)
	userService := users.NewUserService(userRepo, cfg)
	userHandler := users.NewUserHandler(userService, cfg)

	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", userHandler.RegisterUser)
		auth.POST("/login", userHandler.Login)
	}

	users := r.Group("/api/v1/users")
	users.Use(internal.JWTAuthMiddleware(&cfg))
	{
		users.GET("/me", userHandler.GetProfile)
	}

	relationShipsRepo := relationships.NewRelationshipsRepo(db)
	relationShipsService := relationships.NewRelationshipsService(relationShipsRepo, userRepo)
	relationshipsHandler := relationships.NewRelationshipsHandler(relationShipsService)

	relationShips := r.Group("/api/v1/relationships")
	relationShips.Use(internal.JWTAuthMiddleware(&cfg))
	{
		relationShips.POST("/friend-requests", relationshipsHandler.CreateRelationship)
		relationShips.GET("", relationshipsHandler.GetAllRelationships)
		relationShips.PATCH("/users/:target_user_id/friend-requests", relationshipsHandler.AcceptFriendRequest)
		relationShips.DELETE("/users/:target_user_id/friend-requests", relationshipsHandler.CancelOrDeclineFriendRequest)
		relationShips.DELETE("/users/:target_user_id/friends", relationshipsHandler.RemoveFriend)
	}

	channelsRepo := channels.NewChannelsRepo(db)
	channelsService := channels.NewChannelsService(channelsRepo)
	channelsHandler := channels.NewChannelsHandler(channelsService)

	dmChannels := r.Group("/api/v1/users")
	dmChannels.Use(internal.JWTAuthMiddleware(&cfg))
	{
		dmChannels.GET("/:target_user_id/dm", channelsHandler.GetDMChannel)
	}

	channels := r.Group("/api/v1/channels")
	channels.Use(internal.JWTAuthMiddleware(&cfg))
	{
		channels.POST("/groups", channelsHandler.CreateGroupChannel)
		channels.GET("", channelsHandler.GetAllChannels)
	}

}

func (a *App) Shutdown(ctx context.Context) error {
	// First shutdown the HTTP server
	if err := a.HttpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	// Close database connection
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("database connection close: %w", err)
	}

	return nil
}

func (a *App) Close() error {
	return a.db.Close()
}

func handleHealth(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		err := db.PingContext(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"error":  "database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
