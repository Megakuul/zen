// package httplambda provides operations to convert net/http requests and responses to lambda function url events.
package httplambda

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// Requestor provides operations to convert lambda request events to http.Requests.
type Requestor struct{}

func NewRequestor() *Requestor {
	return &Requestor{}
}

func (r *Requestor) Request(ctx context.Context, e events.LambdaFunctionURLRequest) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, e.RequestContext.HTTP.Method, e.RawPath, strings.NewReader(e.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %v", err)
	}
	for k, v := range e.Headers {
		request.Header.Add(k, v)
	}
	return request, nil
}

// Responder implements http.ResponseWriter to convert an http response to a lambda response event.
type Responder struct {
	status int
	body   bytes.Buffer
	header http.Header
}

func NewResponder() *Responder {
	return &Responder{
		status: 0,
		body:   bytes.Buffer{},
		header: http.Header{},
	}
}

func (r *Responder) Header() http.Header {
	return r.header
}

func (r *Responder) Write(input []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.body.Write(input)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (l *Responder) WriteHeader(statusCode int) {
	l.status = statusCode
}

func (l *Responder) Response() events.LambdaFunctionURLResponse {
	resp := events.LambdaFunctionURLResponse{
		StatusCode:      l.status,
		Headers:         map[string]string{},
		Cookies:         []string{},
		Body:            base64.StdEncoding.EncodeToString(l.body.Bytes()),
		IsBase64Encoded: true,
	}
	for k, v := range l.header {
		if len(v) > 0 {
			resp.Headers[k] = v[0]
		}
	}
	return resp
}
