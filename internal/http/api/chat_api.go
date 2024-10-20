package api

import (
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/chatApp/internal/http/controller"
	"github.com/chatApp/internal/pkg/config"
)

type ChatApi struct {
	cfg                 config.ChatApiConfig
	UserController      controller.UserController
	PersonnelController controller.PersonnelController
	ProspectaContoller  controller.ProspectaContoller
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
func NewChatApi(cfg config.ChatApiConfig, pr controller.PersonnelController, psr controller.ProspectaContoller, uc controller.UserController) *ChatApi {
	return &ChatApi{
		cfg:                 cfg,
		UserController:      uc,
		PersonnelController: pr,
		ProspectaContoller:  psr,
	}
}

func (b ChatApi) SetupRoutes(e *echo.Echo) {
	apiV1 := e.Group("/api/v1")

	auth := echojwt.JWT([]byte(b.cfg.AuthSecret))

	userApi := apiV1.Group("/users")
	userApi.POST("", b.UserController.RegisterUser)
	userApi.POST("/login", b.UserController.Login)
	secureUserApi := apiV1.Group("/users")
	secureUserApi.Use(auth)
	secureUserApi.GET("/:id", b.UserController.FindByID)
	secureUserApi.GET("/:username", b.UserController.FindByUserName)

	personnelApi := apiV1.Group("/personnel")
	personnelApi.Use(auth)
	personnelApi.GET("/:id", b.PersonnelController.FindPersonnelByID)
	personnelApi.POST("/filter", b.PersonnelController.Filter)
	personnelApi.POST("", b.PersonnelController.CreatePersonnel)
	personnelApi.PUT("/:id", b.PersonnelController.UpdatePersonnel)
	personnelApi.DELETE("/:id", b.PersonnelController.DeletePersonnel)

	productApi := apiV1.Group("/products")
	productApi.GET("/category/:cat", b.ProspectaContoller.GetProduct)
	productApi.POST("", b.ProspectaContoller.CreateProduct)
}
