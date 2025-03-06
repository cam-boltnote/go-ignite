package routes

import (
	"github.com/cam-boltnote/go-ignite/internal/middleware"
	"github.com/cam-boltnote/go-ignite/internal/services"

	"github.com/gin-gonic/gin"
)

// TestRoutes handles test-related routes
type TestRoutes struct {
	testService *services.TestService
}

// NewTestRoutes creates a new test routes instance
func NewTestRoutes(testService *services.TestService) *TestRoutes {
	return &TestRoutes{
		testService: testService,
	}
}

// RegisterRoutes registers test-related routes
func (r *TestRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	test := rg.Group("/test")
	{
		test.OPTIONS("", middleware.CorsOptionsHandler)
		test.GET("", r.GetTestMessage)
	}
}

// GetTestMessage godoc
// @Summary      Get test message
// @Description  Returns a simple test message to verify the API is working
// @Tags         test
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /test [get]
func (r *TestRoutes) GetTestMessage(c *gin.Context) {
	message := r.testService.GetTestMessage()
	c.JSON(200, gin.H{
		"message": message,
	})
}
