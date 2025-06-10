package core

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/Dev-Umb/go-pkg/errno"
	"github.com/Dev-Umb/go-pkg/logger"

	"github.com/gin-gonic/gin"
	"go.elastic.co/apm"
)

// TraceKey 用于在上下文中存储 trace id 的键
const (
	TraceIDKey   = "traceID"
	StartTimeKey = "startTime"
)

type Response struct {
	Code int `json:"code"`
	//Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	I18n    interface{} `json:"i18n,omitempty"`
	TraceID string      `json:"trace_id"`
}

type Data struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// GetTraceID 从上下文中获取traceID
func GetTraceID(ctx context.Context) string {
	return logger.GetTraceID(ctx)
}

// SetTraceID 将traceID设置到上下文中
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return logger.SetTraceID(ctx, traceID)
}

// GetTraceIDFromGin 从gin上下文中获取traceID
func GetTraceIDFromGin(c *gin.Context) string {
	if c == nil {
		return ""
	}
	// 首先尝试从gin上下文中获取
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	// 如果gin上下文中没有，则尝试从请求上下文中获取
	return GetTraceID(c.Request.Context())
}

func SendResponse(c *gin.Context, err error, data interface{}) {
	// 获取traceID，优先从gin上下文中获取
	traceID := GetTraceIDFromGin(c)

	// 如果上下文中没有traceID，再尝试从apm中获取
	if traceID == "" {
		tx := apm.TransactionFromContext(c.Request.Context())
		if tx != nil {
			traceID = tx.TraceContext().Trace.String()
		}
	}

	// 计算请求耗时
	var elapsed time.Duration
	if startTime, exists := c.Get(StartTimeKey); exists {
		if st, ok := startTime.(time.Time); ok {
			elapsed = time.Since(st)
		}
	}

	code, _, message := errno.DecodeErr(err)

	// 构建响应对象
	response := Response{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Data:    data,
		I18n:    "",
	}

	// 记录响应信息
	responseBytes, _ := json.Marshal(response)

	// 记录详细的响应日志
	if err != nil {
		// 如果有错误，记录错误信息
		logger.Infof(c.Request.Context(), "[%s] Response error: code=%d, message=%s, elapsed=%v",
			traceID, code, message, elapsed)
	} else {
		// 正常响应
		logger.Infof(c.Request.Context(), "[%s] Response completed: code=%d, elapsed=%v, data=%s",
			traceID, code, elapsed, string(responseBytes))
	}

	// 发送响应
	c.JSON(http.StatusOK, response)
}

//func SendBasicResponse(c *gin.Context, err error)  {
//	code, message := errno.DecodeErr(err)
//
//	c.JSON(http.StatusOK, Response{
//		Code:    code,
//		Message: message,
//	})
//}
