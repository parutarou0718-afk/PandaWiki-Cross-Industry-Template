package middleware

import (
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/chaitin/panda-wiki/domain"
	"github.com/chaitin/panda-wiki/log"
)

type ReadOnlyMiddleware struct {
	logger *log.Logger
}

func NewReadonlyMiddleware(logger *log.Logger) *ReadOnlyMiddleware {
	return &ReadOnlyMiddleware{
		logger: logger.WithModule("middleware.readonly"),
	}
}

// echo read only middleware, if request method is not get, return 403 forbidden
func (readonly *ReadOnlyMiddleware) ReadOnly(next echo.HandlerFunc) echo.HandlerFunc {
	readonlyMode := os.Getenv("READONLY") == "1" || strings.ToLower(os.Getenv("READONLY")) == "true"
	return func(c echo.Context) error {
		if !readonlyMode {
			return next(c)
		}
		path := c.Request().URL.Path
		// only check /api/v1 path
		if strings.HasPrefix(path, "/api/v1") {
			method := c.Request().Method
			// skip get
			// skip /api/v1/user/login
			if !isReadOnlyMethod(method) && path != "/api/v1/user/login" {
				readonly.logger.Warn("readonly mode rejected request",
					"method", method,
					"path", path)
				return c.JSON(503, domain.PWResponse{
					Success: false,
					Message: "API is in read-only mode",
				})
			}
		}
		return next(c)
	}
}

func isReadOnlyMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}
