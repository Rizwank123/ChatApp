package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/http/controller"
	"github.com/chatApp/internal/pkg/config"
)

type ChatApi struct {
	cfg            config.ChatApiConfig
	UserController controller.UserController
}

// NewChatApi creates a new ChatApi instance
//
//	@title						Chat API
//	@version					1.0
//	@description				Chat application's set of APIs
//	@termsOfService				https://example.com/terms
//	@contact.name				Mohammad Developer
//	@contact.url				https://example.com
//	@contact.email				mohammad.developer@example.com
//	@host						localhost:8080
//	@BasePath					/api/v1
//	@schemes					http https
//	@securityDefinitions.apiKey	JWT
//	@in							header
//	@name						Authorization
func NewChatApi(cfg config.ChatApiConfig, uc controller.UserController) *ChatApi {
	return &ChatApi{
		cfg:            cfg,
		UserController: uc,
	}
}

func (b ChatApi) SetupRoutes(e *echo.Echo) {
	apiV1 := e.Group("/api/v1")

	auth := echojwt.JWT([]byte(b.cfg.AuthSecret))

	userApi := apiV1.Group("/user")
	userApi.POST("/", b.UserController.RegisterUser)
	//user.POST("/login", b.UserController.LoginUser)
	secureUserApi := apiV1.Group("/user")
	secureUserApi.Use(auth)
	secureUserApi.GET("/:id", b.UserController.FindByID)
}
