package proxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/devopsfaith/krakend/config"
	"github.com/devopsfaith/krakend/transport/http/client/graphql"
)

// NewGraphQLMiddleware returns a middleware with or without the GraphQL
// proxy wrapping the next element (depending on the configuration).
// It supports both queries and mutations.
// For queries, it completes the variables object using the request params.
// For mutations, it overides the defined variables with the request body.
// The resulting request will have a proper graphql body with the query and the
// variables
func NewGraphQLMiddleware(remote *config.Backend) Middleware {
	opt, err := graphql.GetOptions(remote.ExtraConfig)
	if err != nil {
		return EmptyMiddleware
	}

	extractor := graphql.New(*opt)
	var generateBodyFn func(*Request) ([]byte, error)

	switch opt.Type {
	case graphql.OperationMutation:
		f := extractor.BodyExtractor
		generateBodyFn = func(req *Request) ([]byte, error) {
			if req.Body == nil {
				return f(strings.NewReader(""))
			}
			defer req.Body.Close()
			return f(req.Body)
		}

	case graphql.OperationQuery:
		f := extractor.ParamExtractor
		generateBodyFn = func(req *Request) ([]byte, error) {
			return f(req.Params)
		}

	default:
		return EmptyMiddleware
	}

	return func(next ...Proxy) Proxy {
		if len(next) > 1 {
			panic(ErrTooManyProxies)
		}
		return func(ctx context.Context, req *Request) (*Response, error) {
			b, err := generateBodyFn(req)
			if err != nil {
				return nil, err
			}
			req.Body = ioutil.NopCloser(bytes.NewReader(b))
			req.Method = http.MethodPost
			req.Headers["Content-Length"] = []string{strconv.Itoa(len(b))}
			return next[0](ctx, req)
		}
	}
}
