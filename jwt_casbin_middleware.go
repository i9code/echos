package echos

import (
    `net/http`
    "strings"

    "github.com/casbin/casbin/v2"

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
    if nil == config.Skipper {
        config.Skipper = DefaultJWTCasbinConfig.Skipper
    }

    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            if config.Skipper(c) {
                return next(c)
            }

            if pass, err := config.CheckPermission(c); err == nil && pass {
                return next(c)
            } else if err != nil {
                return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
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

func (jcc *JWTCasbinConfig) CheckPermission(c echo.Context) (bool, error) {
    return jcc.Enforcer.Enforce(jcc.GetUsername(c), c.Request().URL.Path, c.Request().Method)
}
