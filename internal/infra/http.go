package infra

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jakottelaar/relay-backend/config"
	"github.com/jakottelaar/relay-backend/internal"
	"github.com/jakottelaar/relay-backend/internal/auth"
	"github.com/jakottelaar/relay-backend/internal/channels"
	"github.com/jakottelaar/relay-backend/internal/messages"
	"github.com/jakottelaar/relay-backend/internal/relationships"
	"github.com/jakottelaar/relay-backend/internal/supabase"
	"github.com/jakottelaar/relay-backend/internal/websocket"
	"github.com/nats-io/nats.go"
)

type AppDependencies struct {
	AuthService            auth.AuthService
	AuthMiddlewareProvider auth.AuthMiddlewareProvider
	SupabaseClient         supabase.SupabaseClient
}

type App struct {
	HttpServer *http.Server
	config     *config.Config
	db         *sql.DB
	deps       *AppDependencies
	nats       *nats.Conn
}

func NewApp(ctx context.Context, cfg *config.Config, deps *AppDependencies) (*App, error) {
	db, err := initializeDB(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("initialize database: %w", err)
	}

	nats, err := initializeNats(cfg)
	if err != nil {
		return nil, fmt.Errorf("initialize nats: %w", err)
	}

	if deps == nil && cfg.Environment != "test" {
		deps = &AppDependencies{
			AuthService: &auth.SupabaseAuthService{
				JwtSecret: cfg.SupabaseJwtSecret,
			},
			AuthMiddlewareProvider: &auth.SupabaseAuthMiddlewareProvider{
				Config: cfg,
			},
			SupabaseClient: supabase.NewSupabaseClient(cfg.SupabaseUrl, cfg.SupabaseApiKey),
		}
	}

	log.Println("database connection established")

	router := gin.Default()
	router.Use(
		gin.Recovery(),
	)

	registerRoutes(router, db, *cfg, deps, nats)

	log.Println("routes registered")

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &App{
		HttpServer: srv,
		config:     cfg,
		db:         db,
		deps:       deps,
		nats:       nats,
	}, nil
}

func registerRoutes(r *gin.Engine, db *sql.DB, cfg config.Config, deps *AppDependencies, nats *nats.Conn) {

	authMiddleware := deps.AuthMiddlewareProvider.AuthMiddleware()

	r.Use(internal.ErrorHandler())

	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true, // Allow all origins for development purposes
		AllowHeaders:    []string{"Authorization", "Content-Type"},
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		MaxAge:          12 * time.Hour,
	}))

	r.GET("/health", authMiddleware, handleHealth(db))

	supabaseClient := deps.SupabaseClient

	channelsRepo := channels.NewChannelsRepo(db)
	channelsService := channels.NewChannelsService(channelsRepo, supabaseClient)
	channelsHandler := channels.NewChannelsHandler(channelsService)

	dmChannels := r.Group("/api/v1/users")
	dmChannels.Use(authMiddleware)
	{
		dmChannels.GET("/:target_user_id/dm", channelsHandler.GetDMChannel)
	}

	channels := r.Group("/api/v1/channels")
	channels.Use(authMiddleware)
	{
		channels.POST("/groups", channelsHandler.CreateGroupChannel)
		channels.GET("", channelsHandler.GetAllChannels)
	}

	messagesRepo := messages.NewMessagesRepo(db)
	messagesService := messages.NewMessagesService(messagesRepo, channelsService)
	messagesHandler := messages.NewMessagesHandler(messagesService)
	messagesEventHandler := messages.NewMessagesEventHandler(nats, messagesService)
	if err := messagesEventHandler.RegisterHandlers(context.Background()); err != nil {
		log.Fatalf("Failed to register message event handlers: %v", err)
	}

	messages := r.Group("/api/v1/channels/:channel_id/messages")
	messages.Use(authMiddleware)
	{
		messages.POST("", messagesHandler.CreateMessage)
		messages.GET("", messagesHandler.GetMessages)
		messages.PATCH("/:message_id", messagesHandler.UpdateMessage)
		messages.DELETE("/:message_id", messagesHandler.DeleteMessage)
	}

	wsManager := websocket.NewManager(nats)
	wsHandler := websocket.NewWebSocketHandler(wsManager, &cfg)
	wsEventHandler := websocket.NewWebsocketEventHandler(nats, wsManager)
	if err := wsEventHandler.RegisterHandlers(); err != nil {
		log.Fatalf("Failed to register WebSocket event handlers: %v", err)
	}

	r.GET("/ws", wsHandler.HandleWebSocket)

	relationShipsRepo := relationships.NewRelationshipsRepo(db)
	relationShipsService := relationships.NewRelationshipsService(relationShipsRepo, supabaseClient, wsManager)
	relationshipsHandler := relationships.NewRelationshipsHandler(relationShipsService)

	relationShips := r.Group("/api/v1/relationships")
	relationShips.Use(authMiddleware)
	{
		relationShips.POST("/friend-requests", relationshipsHandler.CreateRelationship)
		relationShips.GET("", relationshipsHandler.GetAllRelationships)
		relationShips.PATCH("/users/:target_user_id/friend-requests", relationshipsHandler.AcceptFriendRequest)
		relationShips.DELETE("/users/:target_user_id/friend-requests", relationshipsHandler.CancelOrRejectFriendRequest)
		relationShips.DELETE("/users/:target_user_id/friends", relationshipsHandler.RemoveFriend)
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
