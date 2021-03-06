package v1

import (
	"github.com/andrwnv/event-aggregator/controllers"
	"github.com/andrwnv/event-aggregator/middleware"
	"github.com/gin-gonic/gin"
)

func MakeRouter(
	userCtrl *controllers.UserController,
	eventCtrl *controllers.EventController,
	placeCtrl *controllers.PlaceController,
	authCtrl *controllers.AuthController,
	fileCtrl *controllers.FileController,
	commentsCtrl *controllers.CommentController,
	userStoryCtrl *controllers.UserStoryController,
	likeCtrl *controllers.LikeController,
	searchCtrl *controllers.SearchController) *gin.Engine {

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(middleware.CORSMiddleware())

	v1Group := engine.Group("/api/v1")
	{
		userCtrl.MakeRoutesV1(v1Group)
		eventCtrl.MakeRoutesV1(v1Group)
		placeCtrl.MakeRoutesV1(v1Group)
		authCtrl.MakeRoutesV1(v1Group)
		fileCtrl.MakeRoutesV1(v1Group)
		commentsCtrl.MakeRoutesV1(v1Group)
		userStoryCtrl.MakeRoutesV1(v1Group)
		likeCtrl.MakeRoutesV1(v1Group)
		searchCtrl.MakeRoutesV1(v1Group)
	}

	engine.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(404, "Not found")
	})

	engine.NoMethod(func(c *gin.Context) {
		c.AbortWithStatusJSON(405, "Not allowed")
	})

	return engine
}
