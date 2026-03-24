package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Message: message, Data: data})
}

func OKPaginated(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Message: message, Data: data, Meta: meta})
}

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusUnauthorized, APIResponse{Success: false, Error: err})
}

func Forbidden(c *gin.Context, err string) {
	c.JSON(http.StatusForbidden, APIResponse{Success: false, Error: err})
}

func NotFound(c *gin.Context, err string) {
	c.JSON(http.StatusNotFound, APIResponse{Success: false, Error: err})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, APIResponse{Success: false, Error: err})
}

func InternalError(c *gin.Context, err string) {
	c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err})
}