package echos

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo"
)

const (
	defaultIndent = "  "
)

type CustomContext struct {
	echo.Context
}

func (c *CustomContext) writeContentType(value string) {
	header := c.Response().Header()
	if header.Get(echo.HeaderContentType) == "" {
		header.Set(echo.HeaderContentType, value)
	}
}

func (c *CustomContext) jsonPBlob(code int, callback string, i interface{}) (err error) {
	enc := jsoniter.NewEncoder(c.Response())
	_, pretty := c.QueryParams()["pretty"]
	if c.Echo().Debug || pretty {
		enc.SetIndent("", "  ")
	}
	c.writeContentType(echo.MIMEApplicationJavaScriptCharsetUTF8)
	c.Response().WriteHeader(code)
	if _, err = c.Response().Write([]byte(callback + "(")); err != nil {
		return
	}
	if err = enc.Encode(i); err != nil {
		return
	}
	if _, err = c.Response().Write([]byte(");")); err != nil {
		return
	}
	return
}

func (c *CustomContext) json(code int, i interface{}, indent string) error {
	enc := jsoniter.NewEncoder(c.Response())
	if indent != "" {
		enc.SetIndent("", indent)
	}
	c.writeContentType(echo.MIMEApplicationJSONCharsetUTF8)
	c.Response().WriteHeader(code)
	return enc.Encode(i)
}

func (c *CustomContext) JSON(code int, i interface{}) (err error) {
	indent := ""
	if _, pretty := c.QueryParams()["pretty"]; c.Echo().Debug || pretty {
		indent = defaultIndent
	}
	return c.json(code, i, indent)
}

func (c *CustomContext) JSONPretty(code int, i interface{}, indent string) (err error) {
	return c.json(code, i, indent)
}

func (c *CustomContext) JSONBlob(code int, b []byte) (err error) {
	return c.Blob(code, echo.MIMEApplicationJSONCharsetUTF8, b)
}

func (c *CustomContext) JSONP(code int, callback string, i interface{}) (err error) {
	return c.jsonPBlob(code, callback, i)
}

func (c *CustomContext) JSONPBlob(code int, callback string, b []byte) (err error) {
	c.writeContentType(echo.MIMEApplicationJavaScriptCharsetUTF8)
	c.Response().WriteHeader(code)
	if _, err = c.Response().Write([]byte(callback + "(")); err != nil {
		return
	}
	if _, err = c.Response().Write(b); err != nil {
		return
	}
	_, err = c.Response().Write([]byte(");"))
	return
}
