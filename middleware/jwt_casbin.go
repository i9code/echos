package middleware

import (
	"strings"

	"github.com/casbin/casbin"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type (
	Config struct {
		Skipper    middleware.Skipper
		Enforcer   *casbin.Enforcer
		DataSource DataSource
	}
)

var (
	DefaultConfig = Config{
		Skipper: middleware.DefaultSkipper,
	}
)

type DataSource interface {
	GetUsernameByToken(token string) string
}

func JWTMiddleware(e *casbin.Enforcer, ds DataSource) echo.MiddlewareFunc {
	c := DefaultConfig
	c.Enforcer = e
	c.DataSource = ds

	return JWTCasbinWithConfig(c)
}

func JWTCasbinWithConfig(config Config) echo.MiddlewareFunc {
	if config.Skipper == nil {
		config.Skipper = DefaultConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) || config.CheckPermission(c) {
				return next(c)
			}
			return echo.ErrForbidden
		}
	}
}

func (a *Config) GetUsername(c echo.Context) string {
	token := c.Request().Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer")
	if 2 != len(splitToken) {
		return ""
	}
	token = strings.TrimSpace(splitToken[1])

	return a.DataSource.GetUsernameByToken(token)
}

func (a *Config) CheckPermission(c echo.Context) bool {
	return a.Enforcer.Enforce(a.GetUsername(c), c.Request().URL.Path, c.Request().Method)
}
