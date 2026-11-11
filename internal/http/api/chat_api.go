package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/Rizwank123/chatApp/internal/http/controller"
	"github.com/Rizwank123/chatApp/internal/pkg/config"
)

type ChatApi struct {
	cfg                 config.ChatApiConfig
	ChatController      controller.ChatController
	UserController      controller.UserController
	PersonnelController controller.PersonnelController
	PostController      controller.PostController
	SocialController    *controller.SocialController
	MediaController     *controller.MediaController
	SearchController    *controller.SearchController
}

func NewChatApi(
	cfg config.ChatApiConfig,
	cc controller.ChatController,
	pr controller.PersonnelController,
	psc controller.PostController,
	uc controller.UserController,
) *ChatApi {
	return &ChatApi{
		cfg:                 cfg,
		ChatController:      cc,
		UserController:      uc,
		PersonnelController: pr,
		PostController:      psc,
	}
}

// SetSocialController configures the enhanced social controller.
func (b *ChatApi) SetSocialController(sc controller.SocialController) {
	b.SocialController = &sc
}

// SetMediaController configures the media upload controller.
func (b *ChatApi) SetMediaController(mc controller.MediaController) {
	b.MediaController = &mc
}

// SetSearchController configures the search controller.
func (b *ChatApi) SetSearchController(sc controller.SearchController) {
	b.SearchController = &sc
}

func (b ChatApi) SetupRoutes(e *echo.Echo) {
	apiV1 := e.Group("/api/v1")

	auth := echojwt.JWT([]byte(b.cfg.AuthSecret))

	userApi := apiV1.Group("/users")
	userApi.POST("", b.UserController.RegisterUser)
	userApi.POST("/login", b.UserController.Login)
	userApi.GET("/check-username", b.UserController.CheckUsernameAvailability)
	secureUserApi := apiV1.Group("/users")
	secureUserApi.Use(auth)
	secureUserApi.GET("/:id/posts", b.PostController.GetUserPosts)
	secureUserApi.POST("/:id/follow", b.PostController.FollowUser)
	secureUserApi.POST("/:id/unfollow", b.PostController.UnfollowUser)
	secureUserApi.GET("/:id", b.UserController.FindByID)
	secureUserApi.GET("/:username", b.UserController.FindByUserName)

	personnelApi := apiV1.Group("/personnel")
	personnelApi.Use(auth)
	personnelApi.GET("/:id", b.PersonnelController.FindPersonnelByID)
	personnelApi.POST("/filter", b.PersonnelController.Filter)
	personnelApi.POST("", b.PersonnelController.CreatePersonnel)
	personnelApi.PUT("/:id", b.PersonnelController.UpdatePersonnel)
	personnelApi.DELETE("/:id", b.PersonnelController.DeletePersonnel)

	messageApi := apiV1.Group("/messages")
	messageApi.Use(auth)
	messageApi.POST("", b.ChatController.SendMessage)
	messageApi.GET("/:conversationId", b.ChatController.GetMessages)
	messageApi.PUT("/:messageId", b.ChatController.UpdateMessageStatus)

	friendApi := apiV1.Group("/friends")
	friendApi.Use(auth)
	friendApi.POST("", b.ChatController.SendFriendRequest)
	friendApi.POST("/:requestId/accept", b.ChatController.AcceptFriendRequest)
	friendApi.GET("", b.ChatController.GetFriendships)

	conversationApi := apiV1.Group("/conversations")
	conversationApi.Use(auth)
	conversationApi.POST("", b.ChatController.CreateConversation)
	conversationApi.GET("/:conversationId", b.ChatController.GetConversation)
	conversationApi.GET("", b.ChatController.GetUserConversations)

	postApi := apiV1.Group("/posts")
	postApi.Use(auth)
	postApi.POST("", b.PostController.CreatePost)
	postApi.GET("", b.PostController.GetFeedPosts)
	postApi.GET("/explore", b.PostController.GetPublicPosts)

	// If SocialController is present, attach hybrid feed, blocks, saved posts, notifications
	if b.SocialController != nil {
		postApi.GET("/feed", b.SocialController.GetFeed)
		postApi.POST("/:id/save", b.SocialController.SavePost)
		postApi.DELETE("/:id/save", b.SocialController.UnsavePost)

		socialApi := apiV1.Group("/social")
		socialApi.Use(auth)
		socialApi.POST("/follow/:id", b.SocialController.FollowUser)
		socialApi.DELETE("/follow/:id", b.SocialController.UnfollowUser)
		socialApi.POST("/block/:id", b.SocialController.BlockUser)
		socialApi.DELETE("/block/:id", b.SocialController.UnblockUser)
		socialApi.GET("/relationship/:id", b.SocialController.GetRelationship)
		socialApi.GET("/can-chat/:id", b.SocialController.CanChat)
		socialApi.GET("/followers/:id", b.SocialController.GetFollowers)
		socialApi.GET("/following/:id", b.SocialController.GetFollowing)
		socialApi.GET("/search", b.SocialController.SearchUsers)
		socialApi.GET("/counters/:id", b.SocialController.GetUserCounters)
		socialApi.GET("/counters", b.SocialController.GetUserCounters)

		notifApi := apiV1.Group("/notifications")
		notifApi.Use(auth)
		notifApi.GET("", b.SocialController.GetNotifications)
		notifApi.PATCH("/:id/read", b.SocialController.MarkNotificationRead)
	} else {
		postApi.GET("/feed", b.PostController.GetFeedPosts)
	}

	postApi.GET("/:id", b.PostController.GetPost)
	postApi.PUT("/:id", b.PostController.UpdatePost)
	postApi.DELETE("/:id", b.PostController.DeletePost)
	postApi.POST("/:id/like", b.PostController.LikePost)
	postApi.POST("/:id/unlike", b.PostController.UnlikePost)
	postApi.POST("/:id/view", b.PostController.RecordView)
	postApi.POST("/:id/share", b.PostController.RecordShare)
	postApi.POST("/:id/comments", b.PostController.CreateComment)
	postApi.GET("/:id/comments", b.PostController.GetPostComments)
	postApi.DELETE("/:id/comments/:commentId", b.PostController.DeleteComment)

	if b.MediaController != nil {
		mediaApi := apiV1.Group("/media")
		mediaApi.Use(auth)
		mediaApi.POST("/uploads", b.MediaController.InitiateUpload)
		mediaApi.POST("/uploads/:id/complete", b.MediaController.CompleteUpload)
		mediaApi.GET("/:id", b.MediaController.GetMedia)
	}

	if b.SearchController != nil {
		searchApi := apiV1.Group("/search")
		searchApi.Use(auth)
		searchApi.GET("", b.SearchController.UnifiedSearch)
		searchApi.GET("/users", b.SearchController.SearchUsers)
		searchApi.GET("/posts", b.SearchController.SearchPosts)
		searchApi.GET("/hashtags", b.SearchController.SearchHashtags)

		hashtagApi := apiV1.Group("/hashtags")
		hashtagApi.Use(auth)
		hashtagApi.GET("/:tag/posts", b.SearchController.GetHashtagPosts)
		hashtagApi.GET("/:tag", b.SearchController.GetHashtagDetails)
	}

	e.GET("/ws", b.ChatController.StartMessaging)
}
