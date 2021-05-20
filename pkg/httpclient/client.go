package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	httpURL "net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/switch-li/juice/transport/http/middleware/trace"
)

const (
	// DefaultTTL 一次http请求最长执行1分钟
	DefaultTTL = time.Minute
)

// Get get 请求
func Get(url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	return withoutBody(http.MethodGet, url, form, options...)
}

// Delete delete 请求
func Delete(url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	return withoutBody(http.MethodDelete, url, form, options...)
}

func withoutBody(method, url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}

	if len(form) > 0 {
		if url, err = addFormValuesIntoURL(url, form); err != nil {
			return
		}
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/x-www-form-urlencoded; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
		}
	}

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int

	defer func() {
		if opt.alarmObject == nil {
			return
		}

		if opt.alarmVerify != nil && !opt.alarmVerify(body) && err == nil {
			return
		}

		info := &struct {
			TraceID string `json:"trace_id"`
			Request struct {
				Method string `json:"method"`
				URL    string `json:"url"`
			} `json:"request"`
			Response struct {
				HTTPCode int    `json:"http_code"`
				Body     string `json:"body"`
			} `json:"response"`
			Error string `json:"error"`
		}{}

		if opt.trace != nil {
			info.TraceID = opt.trace.ID()
		}
		info.Request.Method = method
		info.Request.URL = url
		info.Response.HTTPCode = httpCode
		info.Response.Body = string(body)
		info.Error = ""
		if err != nil {
			info.Error = fmt.Sprintf("%+v", err)
		}

		raw, _ := json.MarshalIndent(info, "", " ")
		onFailedAlarm(opt.alarmTitle, raw, opt.logger, opt.alarmObject)

	}()

	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, method, url, nil, opt)
		if shouldRetry(ctx, httpCode) || (opt.retryVerify != nil && opt.retryVerify(body)) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

// PostForm post form 请求
func PostForm(url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	return withFormBody(http.MethodPost, url, form, options...)
}

func PostFormMultipart(url, key, file string, form map[string]string, options ...Option) (body []byte, err error) {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fw, err := bodyWriter.CreateFormFile(key, filepath.Base(file))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(fw, f)
	if err != nil {
		return nil, err
	}

	f.Close()

	for k, v := range form {
		bodyWriter.WriteField(k, v)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	if resp, err := http.Post(url, contentType, bodyBuffer); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		return ioutil.ReadAll(resp.Body)
	}
}

func PostFormMultiparts(url, key string, files []string, form map[string]string, options ...Option) (body []byte, err error) {
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	for _, file := range files {
		fw, err := bodyWriter.CreateFormFile(key, filepath.Base(file))
		if err != nil {
			return nil, err
		}

		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fw, f)
		if err != nil {
			return nil, err
		}

		f.Close()
	}

	for k, v := range form {
		bodyWriter.WriteField(k, v)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	if resp, err := http.Post(url, contentType, bodyBuffer); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()

		return ioutil.ReadAll(resp.Body)
	}
}

// PostJSON post json 请求
func PostJSON(url string, raw json.RawMessage, options ...Option) (body []byte, err error) {
	return withJSONBody(http.MethodPost, url, raw, options...)
}

// PutForm put form 请求
func PutForm(url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	return withFormBody(http.MethodPut, url, form, options...)
}

// PutJSON put json 请求
func PutJSON(url string, raw json.RawMessage, options ...Option) (body []byte, err error) {
	return withJSONBody(http.MethodPut, url, raw, options...)
}

// PatchFrom patch form 请求
func PatchFrom(url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	return withFormBody(http.MethodPatch, url, form, options...)
}

// PatchJSON patch json 请求
func PatchJSON(url string, raw json.RawMessage, options ...Option) (body []byte, err error) {
	return withJSONBody(http.MethodPatch, url, raw, options...)
}

func withFormBody(method, url string, form httpURL.Values, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}
	if len(form) == 0 {
		return nil, errors.New("form required")
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/x-www-form-urlencoded; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	formValue := form.Encode()
	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
			Body:       formValue,
		}
	}

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int

	defer func() {
		if opt.alarmObject == nil {
			return
		}

		if opt.alarmVerify != nil && !opt.alarmVerify(body) && err == nil {
			return
		}

		info := &struct {
			TraceID string `json:"trace_id"`
			Request struct {
				Method string `json:"method"`
				URL    string `json:"url"`
			} `json:"request"`
			Response struct {
				HTTPCode int    `json:"http_code"`
				Body     string `json:"body"`
			} `json:"response"`
			Error string `json:"error"`
		}{}

		if opt.trace != nil {
			info.TraceID = opt.trace.ID()
		}
		info.Request.Method = method
		info.Request.URL = url
		info.Response.HTTPCode = httpCode
		info.Response.Body = string(body)
		info.Error = ""
		if err != nil {
			info.Error = fmt.Sprintf("%+v", err)
		}

		raw, _ := json.MarshalIndent(info, "", " ")
		onFailedAlarm(opt.alarmTitle, raw, opt.logger, opt.alarmObject)

	}()

	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, method, url, []byte(formValue), opt)
		if shouldRetry(ctx, httpCode) || (opt.retryVerify != nil && opt.retryVerify(body)) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}

func withJSONBody(method, url string, raw json.RawMessage, options ...Option) (body []byte, err error) {
	if url == "" {
		return nil, errors.New("url required")
	}
	if len(raw) == 0 {
		return nil, errors.New("raw required")
	}

	ts := time.Now()

	opt := getOption()
	defer func() {
		if opt.trace != nil {
			opt.dialog.Success = err == nil
			opt.dialog.CostSeconds = time.Since(ts).Seconds()
			opt.trace.AppendDialog(opt.dialog)
		}

		releaseOption(opt)
	}()

	for _, f := range options {
		f(opt)
	}
	opt.header["Content-Type"] = []string{"application/json; charset=utf-8"}
	if opt.trace != nil {
		opt.header[trace.Header] = []string{opt.trace.ID()}
	}

	ttl := opt.ttl
	if ttl <= 0 {
		ttl = DefaultTTL
	}

	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	defer cancel()

	if opt.dialog != nil {
		decodedURL, _ := httpURL.QueryUnescape(url)
		opt.dialog.Request = &trace.Request{
			TTL:        ttl.String(),
			Method:     method,
			DecodedURL: decodedURL,
			Header:     opt.header,
			Body:       string(raw), // TODO unsafe
		}
	}

	retryTimes := opt.retryTimes
	if retryTimes <= 0 {
		retryTimes = DefaultRetryTimes
	}

	retryDelay := opt.retryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}

	var httpCode int

	defer func() {
		if opt.alarmObject == nil {
			return
		}

		if opt.alarmVerify != nil && !opt.alarmVerify(body) && err == nil {
			return
		}

		info := &struct {
			TraceID string `json:"trace_id"`
			Request struct {
				Method string `json:"method"`
				URL    string `json:"url"`
			} `json:"request"`
			Response struct {
				HTTPCode int    `json:"http_code"`
				Body     string `json:"body"`
			} `json:"response"`
			Error string `json:"error"`
		}{}

		if opt.trace != nil {
			info.TraceID = opt.trace.ID()
		}
		info.Request.Method = method
		info.Request.URL = url
		info.Response.HTTPCode = httpCode
		info.Response.Body = string(body)
		info.Error = ""
		if err != nil {
			info.Error = fmt.Sprintf("%+v", err)
		}

		raw, _ := json.MarshalIndent(info, "", " ")
		onFailedAlarm(opt.alarmTitle, raw, opt.logger, opt.alarmObject)

	}()

	for k := 0; k < retryTimes; k++ {
		body, httpCode, err = doHTTP(ctx, method, url, raw, opt)
		if shouldRetry(ctx, httpCode) || (opt.retryVerify != nil && opt.retryVerify(body)) {
			time.Sleep(retryDelay)
			continue
		}

		return
	}
	return
}
