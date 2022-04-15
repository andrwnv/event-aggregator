package v1

import (
	"github.com/andrwnv/event-aggregator/controllers"
	"github.com/gin-gonic/gin"
	"net/http"
)

func SayHello(c *gin.Context) {
	c.JSON(http.StatusOK, "Hello.")
}

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())

	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/say_hello", SayHello)

		userGroup := apiV1.Group("/user")
		{
			userGroup.POST("create", controllers.RegisterUser)
		}
	}

	return r
}
