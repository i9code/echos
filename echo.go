package echos

import (
    "context"
    "fmt"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "time"

    "github.com/go-playground/validator/v10"
    "github.com/labstack/echo"
    "github.com/labstack/echo/middleware"
)

var (
    DefaultEchoConfig = &EchoConfig{
        Ip:           "",
        Port:         1323,
        Validate:     true,
        ErrorHandler: true,
        Init:         nil,
        Routes:       nil,
    }
)

type (
    EchoFunc   func(e *echo.Echo)
    EchoConfig struct {
        Ip           string
        Port         int
        Validate     bool
        ErrorHandler bool
        Init         EchoFunc
        Routes       []EchoFunc
    }
)

func (ec *EchoConfig) Address() string {
    var address string
    if "" != strings.TrimSpace(ec.Ip) {
        address = fmt.Sprintf("%s:%d", ec.Ip, ec.Port)
    } else {
        address = fmt.Sprintf(":%d", ec.Port)
    }

    return address
}

func Start() {
    StartWith(DefaultEchoConfig)
}

func StartWith(ec *EchoConfig) {
    // 创建Echo对象
    e := echo.New()

    if nil != ec.Init {
        ec.Init(e)
    }
    if nil != ec.Routes {
        for _, route := range ec.Routes {
            route(e)
        }
    }

    // 初始化Validator
    if ec.Validate {
        initValidate()
        // 数据验证
        e.Validator = &customValidator{validator: v}
    }

    // 处理错误
    if ec.ErrorHandler {
        e.HTTPErrorHandler = func(err error, c echo.Context) {
            type response struct {
                Msg       string                                 `json:"msg"`
                Validates validator.ValidationErrorsTranslations `json:"validates"`
            }
            rsp := response{}

            code := http.StatusInternalServerError
            switch re := err.(type) {
            case *echo.HTTPError:
                code = re.Code
                rsp.Msg = re.Error()
            case validator.ValidationErrors:
                lang := c.Request().Header.Get("Accept-Language")
                rsp.Validates = i18n(lang, re)
            }

            c.JSON(code, rsp)
            c.Logger().Error(err)
        }
    }

    // 初始化中间件
    e.Pre(middleware.MethodOverride())
    e.Pre(middleware.RemoveTrailingSlash())

    // e.Use(middleware.CSRF())
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.RequestID())

    // 启动Server
    go func() {
        if err := e.Start(ec.Address()); nil != err {
            e.Logger.Fatal(err)
        }
    }()

    // 等待系统退出中断并响应
    quit := make(chan os.Signal)
    signal.Notify(quit, os.Interrupt)
    <-quit
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := e.Shutdown(ctx); nil != err {
        e.Logger.Fatal(err)
    }
}

func Int64Param(c echo.Context, name string) (int64, error) {
    return strconv.ParseInt(c.Param(name), 10, 64)
}

func IntParam(c echo.Context, name string) (int, error) {
    return strconv.Atoi(c.Param(name))
}
