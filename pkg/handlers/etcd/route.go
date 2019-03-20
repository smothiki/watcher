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
type payload struct {
	Key      string
	KeysOnly bool                   `json:"keys_only" query:"keys_only"`
	Value    map[string]interface{} `json:"value" query:"value"`
	Prefix   bool                   `json:"prefix" query:"prefix"`
	Limit    int64                  `json:"limit" query:"limit"`
	Expire   int64                  `json:"expire" query:"-"`
}

// kv data type, standard json format
type kvmap map[string]interface{}

func (h *Handler) AddRoutes(group *echo.Group) {
	group.Use(func() echo.MiddlewareFunc {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(ctx echo.Context) error {
				var p = new(payload)
				p.Key = ctx.Param("*")

				if err := ctx.Bind(p); err != nil {
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				if ctx.Request().Method != "GET" && slice.ContainsString([]string{"/", ""}, p.Key) {
					err := "invalid etcd key"
					return shared.Responder{Status: http.StatusBadRequest, Success: false, Msg: err}.JSON(ctx)
				}

				ctx.Set("payload", p)
				return next(ctx)
			}
		}
	}())

	group.GET(shared.EmptyPath, h.getName)
	group.GET("/keys*", h.getKey)
	group.PUT("/keys*", h.putKey)
	group.DELETE("/keys*", h.delKey)
}

func (h *Handler) getName(ctx echo.Context) error {
	return shared.Responder{Status: http.StatusOK, Success: true, Result: h.Name()}.JSON(ctx)
}

// Get the key list via payload
func (h *Handler) getKey(ctx echo.Context) error {
	p, _ := ctx.Get("payload").(*payload)

	res, err := h.GetKey(p.Key, p.KeysOnly, p.Prefix, p.Limit)
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
	return shared.Responder{Status: http.StatusOK, Success: true, Result: map[string]interface{}{
		"count": res.Count, "kvs": kvs, "more": res.More,
	}}.JSON(ctx)
}

// Update key by payload
func (h *Handler) putKey(ctx echo.Context) error {
	p, _ := ctx.Get("payload").(*payload)

	// always convert the request value to json
	value, _ := json.Marshal(p.Value)

	// store json
	_, err := h.PutKey(p.Key, string(value), p.Expire)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	return shared.Responder{Status: http.StatusOK, Success: true}.JSON(ctx)
}

// Delete key by payload
func (h *Handler) delKey(ctx echo.Context) error {
	p, _ := ctx.Get("payload").(*payload)

	res, err := h.DeleteKey(p.Key, p.Prefix)
	if err != nil {
		return shared.Responder{Status: http.StatusInternalServerError, Success: false, Msg: err}.JSON(ctx)
	}

	result := map[string]interface{}{"deleted": res.Deleted}
	return shared.Responder{Status: http.StatusOK, Success: true, Result: result}.JSON(ctx)
}
