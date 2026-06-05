package resp

type JsonResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}
