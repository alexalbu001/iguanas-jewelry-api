package middleware

import (
	"errors"
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry/internal/customErrors"
	"github.com/gin-gonic/gin"
)

// https://gin-gonic.com/en/docs/examples/error-handling-middleware/
// https://gin-gonic.com/en/blog/news/how-to-build-one-effective-middleware/
func ErrorHandler() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			var httpErr customerrors.HTTPError
			if errors.As(err, &httpErr) {
				c.JSON(httpErr.StatusCode(), gin.H{
					"message": httpErr.Error(),
					"code":    httpErr.Code(),
					"details": httpErr.Details(),
				})
			}
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
	}
}

// func TestHandler() gin.HandlerFunc {
// 	fmt.Println("Test Handler1")
// 	return func(c *gin.Context) {
// 		fmt.Println("Test handler2")
// 		c.Next()
// 		fmt.Println("Test handler3")
// 	}

// }
