package router

import (
	"github.com/gin-gonic/gin"
	"go-chat/internal/api/handler"
	"go-chat/internal/api/middleware"
)

func New(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	groupHandler *handler.GroupHandler,
	messageHandler *handler.MessageHandler,
	wsHandler *handler.WSHandler,
	uploadHandler *handler.UploadHandler,
	friendHandler *handler.FriendHandler,
	uploadDir string,
	jwtSecret string,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	_ = router.SetTrustedProxies(nil)
	router.Static("/uploads", uploadDir)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}

		protected := api.Group("")
		protected.Use(middleware.JWTAuth(jwtSecret))
		{
			protected.GET("/users/me", userHandler.Me)
			protected.PATCH("/users/me", userHandler.UpdateMe)
			protected.GET("/users/:id/profile", userHandler.PublicProfile)
			protected.GET("/users/:id/online", userHandler.Online)
			protected.POST("/users/online/batch", userHandler.OnlineBatch)
			protected.POST("/groups", groupHandler.Create)
			protected.PATCH("/groups/:id", groupHandler.UpdateProfile)
			protected.POST("/groups/:id/join", groupHandler.Join)
			protected.GET("/groups", groupHandler.ListMine)
			protected.GET("/groups/:id/members", groupHandler.ListMembers)
			protected.POST("/groups/:id/leave", groupHandler.Leave)
			protected.DELETE("/groups/:id/members/:user_id", groupHandler.RemoveMember)
			protected.GET("/messages", messageHandler.List)
			protected.GET("/conversations/private", messageHandler.ListPrivateConversations)
			protected.POST("/conversations/read", messageHandler.MarkConversationRead)
			protected.POST("/uploads/images", uploadHandler.UploadImage)

			// 好友模块
			protected.GET("/users/search", friendHandler.SearchUsers)
			protected.POST("/friends/requests", friendHandler.SendRequest)
			protected.GET("/friends/requests/pending", friendHandler.ListPendingRequests)
			protected.PUT("/friends/requests/:id/accept", friendHandler.AcceptRequest)
			protected.PUT("/friends/requests/:id/reject", friendHandler.RejectRequest)
			protected.GET("/friends", friendHandler.ListFriends)
			protected.DELETE("/friends/:id", friendHandler.DeleteFriend)
		}
	}

	router.GET("/ws", wsHandler.Handle)

	return router
}
