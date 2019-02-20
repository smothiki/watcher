package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"k8s.io/apimachinery/pkg/fields"

	"github.com/srelab/watcher/pkg/handlers/shared"

	"github.com/srelab/common/log"
	"github.com/srelab/common/slice"

	"github.com/labstack/echo/middleware"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/g"
)

// Webserver validation function
type Validator struct {
	validate *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// Implement the bind method to verify the request's struct for parameter validation
type BinderWithValidation struct{}

func (b BinderWithValidation) Bind(i interface{}, ctx echo.Context) error {
	binder := &echo.DefaultBinder{}

	var body []byte
	if ctx.Request().Body != nil {
		body, _ = ioutil.ReadAll(ctx.Request().Body)
	}

	// NopCloser returns a ReadCloser with a no-op Close method wrapping the provided Reader r.
	// Enables ctx to bind request data multiple times
	// TODO: When there is a performance hazard in the future, you can consider parameterizing it.
	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if err := binder.Bind(i, ctx); err != nil {
		return errors.New(err.(*echo.HTTPError).Message.(string))
	}

	if err := ctx.Validate(i); err != nil {
		var buf bytes.Buffer

		for _, fieldErr := range err.(validator.ValidationErrors) {
			buf.WriteString("Validation failed on ")
			buf.WriteString(fieldErr.Tag())
			buf.WriteString(" for ")
			buf.WriteString(fieldErr.StructField())
			buf.WriteString("\n")
		}

		return errors.New(buf.String())
	}

	ctx.Request().Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return nil
}

func NewHandlersEngine() *echo.Echo {
	engine := echo.New()
	logger := log.With("handlers", "engine")

	// Turn on middleware and customize the output of the log to match the log output of the watcher service
	engine.Use(middleware.CORS(), middleware.Recover(), middleware.RequestID(), func() echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				req := c.Request()
				res := c.Response()

				var err error
				if err = next(c); err != nil {
					c.Error(err)
				}

				id := req.Header.Get(echo.HeaderXRequestID)
				if id == "" {
					id = res.Header().Get(echo.HeaderXRequestID)
				}
				reqSize := req.Header.Get(echo.HeaderContentLength)
				if reqSize == "" {
					reqSize = "0"
				}

				msg := fmt.Sprintf("[%s][%s][%s][%s][%s][%d][%s]",
					id, c.RealIP(), req.Host, req.Method, req.RequestURI, res.Status, req.UserAgent(),
				)

				if slice.ContainsInt([]int{http.StatusOK}, res.Status) {
					logger.Info(msg)
				} else {
					logger.Error(msg)
				}

				return err
			}
		}
	}())

	engine.HidePort = true
	engine.HideBanner = true
	engine.Debug = g.Config().Http.Debug
	engine.Binder = &BinderWithValidation{}

	// Override the default Validator
	engine.Validator = func() echo.Validator {
		v := validator.New()

		_ = v.RegisterValidation("json", func(fl validator.FieldLevel) bool {
			var js json.RawMessage
			return json.Unmarshal([]byte(fl.Field().String()), &js) == nil
		})

		_ = v.RegisterValidation("in", func(fl validator.FieldLevel) bool {
			value := fl.Field().String()
			if slice.ContainsString(strings.Split(fl.Param(), ";"), value) || value == "" {
				return true
			}

			return false
		})

		_ = v.RegisterValidation("k8s_selector", func(fl validator.FieldLevel) bool {
			if _, err := fields.ParseSelector(fl.Field().String()); err != nil {
				return false
			}

			return true
		})

		return &Validator{validate: v}
	}()

	// Override the default error handler
	engine.HTTPErrorHandler = func(err error, ctx echo.Context) {
		var (
			code = http.StatusInternalServerError
			msg  interface{}
		)

		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = he.Message
			if he.Internal != nil {
				err = fmt.Errorf("%v, %v", err, he.Internal)
			}
		} else if engine.Debug {
			msg = err.Error()
		} else {
			msg = http.StatusText(code)
		}

		// Send response
		if !ctx.Response().Committed {
			// https://www.w3.org/Protocols/rfc2616/rfc2616-sec9.html
			if ctx.Request().Method == http.MethodHead {
				err = ctx.NoContent(code)
			} else {
				err = shared.Responder{Status: code, Success: false, Msg: msg}.JSON(ctx)
			}
			if err != nil {
				logger.Error(err)
			}
		}
	}

	return engine
}
