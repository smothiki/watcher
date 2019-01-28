package etcd

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo"
	"github.com/srelab/common/slice"
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
	Key      string
	KeysOnly bool                   `json:"keys_only" query:"keys_only"`
	Value    map[string]interface{} `json:"value" query:"value"`
	Prefix   bool                   `json:"prefix" query:"prefix"`
	Limit    int64                  `json:"limit" query:"limit"`
}

// kv data type, standard json format
type kvmap map[string]interface{}

func (h *Handler) AddRoutes(group *echo.Group) *echo.Group {
	group.Use(func() echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				var payload = new(Payload)
				payload.Key = ctx.Param("*")

				if err := ctx.Bind(payload); err != nil {
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				if ctx.Request().Method != "GET" && slice.ContainsString([]string{"/", ""}, payload.Key) {
					err := "invalid etcd key"
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				ctx.Set("payload", payload)
				return next(ctx)
			}
		}
	}())

	group.GET("", h.getName)
	group.GET("/keys*", h.getKey)
	group.PUT("/keys*", h.putKey)
	group.DELETE("/keys*", h.delKey)
	return group
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

// Get the key list via payload
func (h *Handler) getKey(ctx echo.Context) error {
	payload, _ := ctx.Get("payload").(*Payload)

	res, err := h.eGet(payload.Key, payload.KeysOnly, payload.Prefix, payload.Limit)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	kvs := make([]kvmap, 0)
	for _, ev := range res.Kvs {
		kv := kvmap{}
		value := new(kvmap)

		kv["key"] = string(ev.Key)
		if err := json.Unmarshal(ev.Value, value); err != nil {
			kv["value"] = ev.Value
		} else {
			kv["value"] = value
		}

		kv["mod_revision"] = ev.ModRevision
		kv["create_revision"] = ev.CreateRevision
		kv["version"] = ev.Version

		kvs = append(kvs, kv)
	}

	// more indicates if there are more keys to return in the requested range.
	result := map[string]interface{}{"count": res.Count, "kvs": kvs, "more": res.More}
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
