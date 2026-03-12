package main

import (
	"context"
	"log"
	"os"

	"go-chat/internal/api/handler"
	apiRouter "go-chat/internal/api/router"
	"go-chat/internal/config"
	"go-chat/internal/repository"
	"go-chat/internal/service"
	"go-chat/internal/ws"
)

func main() {
	cfg := config.Load()

	db, err := repository.NewMySQL(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("connect mysql failed: %v", err)
	}

	if err := repository.AutoMigrate(db); err != nil {
		log.Fatalf("auto migrate failed: %v", err)
	}

	redisClient, err := repository.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("connect redis failed, continue without redis: %v", err)
		redisClient = nil
	}

	userRepo := repository.NewUserRepository(db)
	groupRepo := repository.NewGroupRepository(db)
	messageRepo := repository.NewMessageRepository(db)
	friendRepo := repository.NewFriendRepository(db)
	manager := ws.NewManager()

	authService := service.NewAuthService(userRepo, cfg)
	profileService := service.NewProfileService(userRepo)
	onlineService := service.NewOnlineService(userRepo, redisClient, manager)
	groupService := service.NewGroupService(groupRepo)
	chatService := service.NewChatService(messageRepo, groupRepo, userRepo, friendRepo)
	friendService := service.NewFriendService(friendRepo, userRepo)

	if redisClient != nil {
		redisAdapter := ws.NewRedisAdapter(redisClient, manager)
		go redisAdapter.Subscribe(context.Background())
		log.Println("Redis Pub/Sub adapter initialized")
	}

	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(profileService, onlineService)
	groupHandler := handler.NewGroupHandler(groupService)
	messageHandler := handler.NewMessageHandler(chatService)
	wsHandler := handler.NewWSHandler(authService, chatService, manager, redisClient)
	uploadHandler := handler.NewUploadHandler(cfg.UploadDir, cfg.UploadMaxImage)
	friendHandler := handler.NewFriendHandler(friendService, userRepo, manager)

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("create upload dir failed: %v", err)
	}

	router := apiRouter.New(authHandler, userHandler, groupHandler, messageHandler, wsHandler, uploadHandler, friendHandler, cfg.UploadDir, cfg.JWTSecret)

	log.Printf("server started at %s", cfg.ServerAddr)
	if err := router.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
