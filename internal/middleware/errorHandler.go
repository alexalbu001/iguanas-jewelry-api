package middleware

import (
	"errors"
	"net/http"

	customerrors "github.com/alexalbu001/iguanas-jewelry-api/internal/customErrors"
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
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Internal server error",
					"code":    "INTERNAL_SERVER_ERROR",
					"details": "",
				})
				return
			}
		}

	}
}
