package handlers

import (
	"bytes"
	"errors"

	"github.com/go-playground/validator"
	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/event"
	"github.com/srelab/watcher/pkg/g"
)

// Handler is implemented by any handler.
// The Handle method is used to process event
type Handler interface {
	Name() string

	Init(config *g.Configuration) error
	Created(event event.Event)
	Deleted(event event.Event)
	Updated(event event.Event)
}

// Webserver validation function
type Validator struct {
	validate *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	return v.validate.Struct(i)
}

// Implement the bind method to verify the request's struct for parameter validation
type BinderWithValidation struct{}

func (BinderWithValidation) Bind(i interface{}, ctx echo.Context) error {
	binder := &echo.DefaultBinder{}

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

	return nil
}

// Default handler implements Handler interface,
// print each event with JSON format
type DefaultHandler struct {
	Handler
}

func (h *DefaultHandler) Name() string {
	return "default"
}

// Init initializes handler configuration
// Do nothing for default handler
func (h *DefaultHandler) Init(config *g.Configuration) error {
	return nil
}

func (h *DefaultHandler) Created(e event.Event) {
}

func (h *DefaultHandler) Deleted(e event.Event) {
}

func (h *DefaultHandler) Updated(e event.Event) {
}

// Responder in order to unify the returned response structure
type Responder struct {
	Status     int         `json:"-"`
	Success    bool        `json:"success"`
	Result     interface{} `json:"result"`
	Msg        string      `json:"msg"`
	Pagination interface{} `json:"pagination,omitempty"`
}

func (r *Responder) JSON(ctx echo.Context) error {
	return ctx.JSON(r.Status, r)
}
