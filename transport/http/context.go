package http

import (
	"bytes"
	stdctx "context"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/mel2oo/juice/pkg/logger"
	"github.com/mel2oo/juice/pkg/logger/zap"
	"github.com/mel2oo/juice/transport/http/middleware/trace"
)

type HandlerFunc func(c Context)

type Trace = trace.T

const (
	_Alias            = "_alias_"
	_TraceName        = "_trace_"
	_LoggerName       = "_logger_"
	_BodyName         = "_body_"
	_PayloadName      = "_payload_"
	_GraphPayloadName = "_graph_payload_"
	_UserID           = "_user_id_"
	_UserName         = "_user_name_"
	_AbortErrorName   = "_abort_error_"
)

type Context interface {
	init()

	// ShouldBindQuery 反序列化 querystring
	// tag: `form:"xxx"` (注：不要写成 query)
	ShouldBindQuery(obj interface{}) error

	// ShouldBindPostForm 反序列化 postform (querystring会被忽略)
	// tag: `form:"xxx"`
	ShouldBindPostForm(obj interface{}) error

	// ShouldBindForm 同时反序列化 querystring 和 postform;
	// 当 querystring 和 postform 存在相同字段时，postform 优先使用。
	// tag: `form:"xxx"`
	ShouldBindForm(obj interface{}) error

	// ShouldBindJSON 反序列化 postjson
	// tag: `json:"xxx"`
	ShouldBindJSON(obj interface{}) error

	// ShouldBindURI 反序列化 path 参数(如路由路径为 /user/:name)
	// tag: `uri:"xxx"`
	ShouldBindURI(obj interface{}) error

	ShouldBindFormMultipart(obj interface{}) error
	SaveUploadedFile(file *multipart.FileHeader, dst string) error

	// Redirect 重定向
	Redirect(code int, location string)

	// Trace 获取 Trace 对象
	Trace() Trace
	setTrace(trace Trace)
	disableTrace()

	// Logger 获取 Logger 对象
	Logger() logger.Logger
	setLogger(logger logger.Logger)

	// Payload 正确返回
	Payload(payload interface{})
	getPayload() interface{}

	// GraphPayload GraphQL返回值 与 api 返回结构不同
	GraphPayload(payload interface{})
	getGraphPayload() interface{}

	// File 返回文件
	File(filename, filepath string)
	FileData(filepath string)
	FileFromMultipart(filename string, data []byte)

	// AbortWithError 错误返回
	AbortWithError(err Error)
	abortError() Error

	// Header 获取 Header 对象
	Header() http.Header
	// GetHeader 获取 Header
	GetHeader(key string) string
	// SetHeader 设置 Header
	SetHeader(key, value string)

	// UserID 获取 JWT 中 UserID
	UserID() int64
	setUserID(userID int64)

	// UserName 获取 JWT 中 UserName
	UserName() string
	setUserName(userName string)

	// Alias 设置路由别名 for metrics uri
	Alias() string
	setAlias(path string)

	// Request 获取 Request 对象
	Request() *http.Request
	// RawData 获取 Request.Body
	RawData() []byte
	// Method 获取 Request.Method
	Method() string
	// Host 获取 Request.Host
	Host() string
	// Path 获取 请求的路径 Request.URL.Path (不附带 querystring)
	Path() string
	// URI 获取 unescape 后的 Request.URL.RequestURI()
	URI() string
	// RequestContext 获取请求的 context (当 client 关闭后，会自动 canceled)
	RequestContext() StdContext

	// ResponseWriter 获取 ResponseWriter 对象
	ResponseWriter() gin.ResponseWriter
}

var contextPool = &sync.Pool{
	New: func() interface{} {
		return new(context)
	},
}

func newContext(ctx *gin.Context) Context {
	context := contextPool.Get().(*context)
	context.ctx = ctx
	return context
}

func releaseContext(ctx Context) {
	c := ctx.(*context)
	c.ctx = nil
	contextPool.Put(c)
}

var _ Context = (*context)(nil)

type context struct {
	ctx *gin.Context
}

type StdContext struct {
	stdctx.Context
	Trace
	logger.Logger
}

func (c *context) init() {
	body, err := c.ctx.GetRawData()
	if err != nil {
		panic(err)
	}

	c.ctx.Set(_BodyName, body)                                   // cache body是为了trace使用
	c.ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body
}

// ShouldBindQuery 反序列化querystring
// tag: `form:"xxx"` (注：不要写成query)
func (c *context) ShouldBindQuery(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Query)
}

// ShouldBindPostForm 反序列化 postform (querystring 会被忽略)
// tag: `form:"xxx"`
func (c *context) ShouldBindPostForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormPost)
}

// ShouldBindForm 同时反序列化querystring和postform;
// 当querystring和postform存在相同字段时，postform优先使用。
// tag: `form:"xxx"`
func (c *context) ShouldBindForm(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.Form)
}

// ShouldBindJSON 反序列化postjson
// tag: `json:"xxx"`
func (c *context) ShouldBindJSON(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.JSON)
}

// ShouldBindURI 反序列化path参数(如路由路径为 /user/:name)
// tag: `uri:"xxx"`
func (c *context) ShouldBindURI(obj interface{}) error {
	return c.ctx.ShouldBindUri(obj)
}

func (c *context) ShouldBindFormMultipart(obj interface{}) error {
	return c.ctx.ShouldBindWith(obj, binding.FormMultipart)
}

func (c *context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return c.ctx.SaveUploadedFile(file, dst)
}

// Redirect 重定向
func (c *context) Redirect(code int, location string) {
	c.ctx.Redirect(code, location)
}

func (c *context) Trace() Trace {
	t, ok := c.ctx.Get(_TraceName)
	if !ok || t == nil {
		return nil
	}

	return t.(Trace)
}

func (c *context) setTrace(trace Trace) {
	c.ctx.Set(_TraceName, trace)
}

func (c *context) disableTrace() {
	c.setTrace(nil)
}

func (c *context) Logger() logger.Logger {
	logger, ok := c.ctx.Get(_LoggerName)
	if !ok {
		return nil
	}

	return logger.(*zap.ZapLogger)
}

func (c *context) setLogger(logger logger.Logger) {
	c.ctx.Set(_LoggerName, logger)
}

func (c *context) getPayload() interface{} {
	if payload, ok := c.ctx.Get(_PayloadName); ok != false {
		return payload
	}
	return nil
}

func (c *context) Payload(payload interface{}) {
	c.ctx.Set(_PayloadName, payload)
}

func (c *context) getGraphPayload() interface{} {
	if payload, ok := c.ctx.Get(_GraphPayloadName); ok != false {
		return payload
	}
	return nil
}

func (c *context) File(filename, filepath string) {
	fileContentDisposition := "attachment;filename=\"" + filename + "\""
	c.ctx.Header("Content-Type", "multipart/form-data")
	c.ctx.Header("Content-Disposition", fileContentDisposition)
	c.ctx.File(filepath)
}

func (c *context) FileData(filepath string) {
	c.ctx.File(filepath)
}

func (c *context) FileFromMultipart(filename string, data []byte) {
	fileContentDisposition := "attachment;filename=\"" + filename + "\""
	c.ctx.Header("Content-Type", "multipart/form-data")
	c.ctx.Header("Content-Disposition", fileContentDisposition)
	c.ctx.Data(http.StatusOK, "multipart/form-data", data)
}

func (c *context) GraphPayload(payload interface{}) {
	c.ctx.Set(_GraphPayloadName, payload)
}

func (c *context) Header() http.Header {
	header := c.ctx.Request.Header

	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)

		clone[k] = value
	}
	return clone
}

func (c *context) GetHeader(key string) string {
	return c.ctx.GetHeader(key)
}

func (c *context) SetHeader(key, value string) {
	c.ctx.Header(key, value)
}

func (c *context) UserID() int64 {
	val, ok := c.ctx.Get(_UserID)
	if !ok {
		return 0
	}

	return val.(int64)
}

func (c *context) setUserID(userID int64) {
	c.ctx.Set(_UserID, userID)
}

func (c *context) UserName() string {
	val, ok := c.ctx.Get(_UserName)
	if !ok {
		return ""
	}

	return val.(string)
}

func (c *context) setUserName(userName string) {
	c.ctx.Set(_UserName, userName)
}

func (c *context) AbortWithError(err Error) {
	if err != nil {
		httpCode := err.GetHttpCode()
		if httpCode == 0 {
			httpCode = http.StatusInternalServerError
		}

		c.ctx.AbortWithStatus(httpCode)
		c.ctx.Set(_AbortErrorName, err)
	}
}

func (c *context) abortError() Error {
	err, _ := c.ctx.Get(_AbortErrorName)
	return err.(Error)
}

func (c *context) Alias() string {
	path, ok := c.ctx.Get(_Alias)
	if !ok {
		return ""
	}

	return path.(string)
}

func (c *context) setAlias(path string) {
	if path = strings.TrimSpace(path); path != "" {
		c.ctx.Set(_Alias, path)
	}
}

// Request 获取 Request
func (c *context) Request() *http.Request {
	return c.ctx.Request
}

func (c *context) RawData() []byte {
	body, ok := c.ctx.Get(_BodyName)
	if !ok {
		return nil
	}

	return body.([]byte)
}

// Method 请求的method
func (c *context) Method() string {
	return c.ctx.Request.Method
}

// Host 请求的host
func (c *context) Host() string {
	return c.ctx.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *context) Path() string {
	return c.ctx.Request.URL.Path
}

// URI unescape后的uri
func (c *context) URI() string {
	uri, _ := url.QueryUnescape(c.ctx.Request.URL.RequestURI())
	return uri
}

// RequestContext 获取请求的 context (当client关闭后，会自动canceled)
func (c *context) RequestContext() StdContext {
	return StdContext{
		c.ctx.Request.Context(),
		c.Trace(),
		c.Logger(),
	}
}

// ResponseWriter 获取 ResponseWriter
func (c *context) ResponseWriter() gin.ResponseWriter {
	return c.ctx.Writer
}
