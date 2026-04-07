package errcode

type Code struct {
	Value   int
	Message string
}

func (c *Code) WithMessage(msg string) *Code {
	return &Code{Value: c.Value, Message: msg}
}

var (
	OK = &Code{Value: 0, Message: "success"}

	BadRequest   = &Code{Value: 400, Message: "bad request"}
	Unauthorized = &Code{Value: 401, Message: "unauthorized"}
	Forbidden    = &Code{Value: 403, Message: "forbidden"}
	NotFound     = &Code{Value: 404, Message: "not found"}
	Conflict     = &Code{Value: 409, Message: "conflict"}
	Internal     = &Code{Value: 500, Message: "internal server error"}
)
