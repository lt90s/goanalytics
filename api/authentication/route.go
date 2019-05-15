package authentication

import (
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/goanalytics/api/middlewares"
	"github.com/lt90s/goanalytics/event/pubsub"
)

func SetupRoute(router *gin.Engine, jwtMiddleware *jwt.GinJWTMiddleware, adminStore store, publisher pubsub.Publisher) {
	adminGroup := router.Group("/admin")

	authGroup := adminGroup.Group("/auth")

	authGroup.POST("/login", jwtMiddleware.LoginHandler)

	accountGroup := authGroup.Group("/account", jwtMiddleware.MiddlewareFunc())
	requireAdminRole := middlewares.RequireRoleMiddleware([]string{"admin"})

	// get account info
	accountGroup.GET("", getAccountInfoHandler(adminStore))
	// create account
	accountGroup.POST("", requireAdminRole, createAccountHandler(adminStore))
	// modify account info
	accountGroup.PUT("")
	// delete account
	accountGroup.DELETE("", requireAdminRole, deleteAccountHandler(adminStore))

	accountGroup.GET("/all", requireAdminRole, GetAccountInfosHandler(adminStore))


	appGroup := adminGroup.Group("/app", jwtMiddleware.MiddlewareFunc())
	// get apps info
	appGroup.GET("", getAppsHandler(adminStore))
	appGroup.POST("", createAppHandler(adminStore))
	appGroup.DELETE("", deleteAppHandler(adminStore, publisher))
}


