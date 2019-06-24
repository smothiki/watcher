package gateway

type SliceResult struct {
	Status bool          `json:"status"`
	Data   []interface{} `json:"data"`
}

type MapResult struct {
	Status bool        `json:"status"`
	Data   interface{} `json:"data"`
}

type RegisterServicePayload struct {
	Host   string `json:"host" validate:"required"`
	Type   string `json:"type" validate:"required"`
	Port   int    `json:"port" validate:"required"`
	HcPath string `json:"hc_path" validate:"required"`
	HcPort int    `json:"hc_port" validate:"required"`
}

type UnRegisterServicePayload struct {
	Host string `json:"host" validate:"required"`
	Port int    `json:"port" validate:"required"`
}
