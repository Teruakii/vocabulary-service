package response

import "github.com/gin-gonic/gin"

type Envelope struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func OK(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Envelope{Data: data})
}

func Error(c *gin.Context, status int, err error) {
	c.JSON(status, Envelope{Error: err.Error()})
}

func ErrorMsg(c *gin.Context, status int, msg string) {
	c.JSON(status, Envelope{Error: msg})
}

func NoContent(c *gin.Context) {
	c.Status(204)
}
