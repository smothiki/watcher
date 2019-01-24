package etcd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/srelab/watcher/pkg/handlers/shared"
)

// Handle the payload required by etcd,
// TODO: The naming is more standardized, distinguishing different usage scenarios
//
// Key: from the URL
// Value: from Request Body, only used in the put method
// Prefix && limit: This parameter is obtained from querystring when the request method is Get or Delete,
// otherwise it is obtained from request body
type Payload struct {
	Key    string                 `validate:"required"`
	Value  map[string]interface{} `json:"value" query:"value" validate:"-"`
	Prefix bool                   `json:"prefix" query:"prefix" validate:"-"`
	Limit  int64                  `json:"limit" query:"limit" validate:"-"`
}

// kv data type, standard json format
type kvmap map[string]interface{}

func (h *Handler) AddRoutes(group *echo.Group) *echo.Group {
	group.Use(func() echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				var payload = new(Payload)
				payload.Key = strings.TrimLeft(ctx.Param("*"), "/")
				if payload.Key != "" {
					payload.Key = "/" + payload.Key
				}

				if err := ctx.Bind(payload); err != nil {
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				ctx.Set("payload", payload)
				return next(ctx)
			}
		}
	}())

	group.GET("", h.getName)
	group.GET("/keys/*", h.getKey)
	group.PUT("/keys/*", h.putKey)
	group.DELETE("/keys/*", h.delKey)
	return group
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

// Get the key list via payload
func (h *Handler) getKey(ctx echo.Context) error {
	payload, _ := ctx.Get("payload").(*Payload)

	res, err := h.eGet(payload.Key, payload.Prefix, payload.Limit)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	kvs := make([]kvmap, 0)
	for _, ev := range res.Kvs {
		kv := kvmap{}
		value := new(kvmap)

		if err := json.Unmarshal(ev.Value, value); err != nil {
			fmt.Println(err)
			kv[string(ev.Key)] = ev.Value
		} else {
			kv[string(ev.Key)] = value
		}

		kv["mod_revision"] = ev.ModRevision
		kv["create_revision"] = ev.CreateRevision
		kv["version"] = ev.Version

		kvs = append(kvs, kv)
	}

	result := map[string]interface{}{"count": res.Count, "kvs": kvs}
	return shared.Responder{Status: http.StatusOK, Success: true, Result: result}.JSON(ctx)
}

// Update key by payload
func (h *Handler) putKey(ctx echo.Context) error {
	payload, _ := ctx.Get("payload").(*Payload)

	// always convert the request value to json
	value, _ := json.Marshal(payload.Value)

	// store json
	_, err := h.ePut(payload.Key, string(value))
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// Delete key by payload
func (h *Handler) delKey(ctx echo.Context) error {
	payload, _ := ctx.Get("payload").(*Payload)

	res, err := h.eDelete(payload.Key, payload.Prefix)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	result := map[string]interface{}{"deleted": res.Deleted}
	return shared.Responder{Status: http.StatusOK, Success: true, Result: result}.JSON(ctx)
}
