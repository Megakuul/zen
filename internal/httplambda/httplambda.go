// package httplambda provides operations to convert net/http requests and responses to lambda function url events.
package httplambda

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// LambdaRequestor provides operations to convert lambda request events to http.Requests. 
type LambdaRequestor struct { }

func (l *LambdaRequestor) Request(ctx context.Context, e events.LambdaFunctionURLRequest) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, e.RequestContext.HTTP.Method, e.RawPath, strings.NewReader(e.Body))
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %v", err)
	}
	for k, v := range e.Headers {
		request.Header.Add(k, v)
	}
	return request, nil
}

// LambdaResponder implements http.ResponseWriter to convert an http response to a lambda response event.
type LambdaResponder struct {
	status int
	body   strings.Builder
	header http.Header
}

func (l *LambdaResponder) Header() http.Header {
	return l.header
}

func (l *LambdaResponder) Write(input []byte) (int, error) {
	if l.status == 0 {
		l.status = http.StatusOK
	}
	n, err := l.body.Write(input)
	if err != nil {
		return n, err
	}
	return n, nil
}

func (l *LambdaResponder) WriteHeader(statusCode int) {
	l.status = statusCode
}

func (l *LambdaResponder) Response() events.LambdaFunctionURLResponse {
	resp := events.LambdaFunctionURLResponse{
		StatusCode:      l.status,
		Headers:         map[string]string{},
		Cookies:         []string{},
		Body:            l.body.String(),
		IsBase64Encoded: false,
	}
	for k, v := range l.header {
		if len(v) > 0 {
			resp.Headers[k] = v[0]
		}
	}
	return resp
}
