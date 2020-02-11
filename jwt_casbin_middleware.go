package echos

import (
    "strings"

    "github.com/casbin/casbin"

    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
)

type (
    JWTCasbinConfig struct {
        Skipper    middleware.Skipper
        Enforcer   *casbin.Enforcer
        DataSource DataSource
    }
)

var (
    DefaultJWTCasbinConfig = JWTCasbinConfig{
        Skipper: middleware.DefaultSkipper,
    }
)

type DataSource interface {
    GetUsernameByToken(token string) string
}

func JWTCasbinMiddleware(e *casbin.Enforcer, ds DataSource) echo.MiddlewareFunc {
    c := DefaultJWTCasbinConfig
    c.Enforcer = e
    c.DataSource = ds

    return JWTCasbinWithConfig(c)
}

func JWTCasbinWithConfig(config JWTCasbinConfig) echo.MiddlewareFunc {
    if config.Skipper == nil {
        config.Skipper = DefaultJWTCasbinConfig.Skipper
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

func (jcc *JWTCasbinConfig) GetUsername(c echo.Context) string {
    token := c.Request().Header.Get("Authorization")
    splitToken := strings.Split(token, "Bearer")
    if 2 != len(splitToken) {
        return ""
    }
    token = strings.TrimSpace(splitToken[1])

    return jcc.DataSource.GetUsernameByToken(token)
}

func (jcc *JWTCasbinConfig) CheckPermission(c echo.Context) bool {
    return jcc.Enforcer.Enforce(jcc.GetUsername(c), c.Request().URL.Path, c.Request().Method)
}
