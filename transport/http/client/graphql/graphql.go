package graphql

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/devopsfaith/krakend/config"
)

// Namespace is the key for the backend's extra config
const Namespace = "github.com/devopsfaith/krakend/transport/http/client/graphql"

// OperationType contains all the operations allowed by graphql
type OperationType string

const (
	// OperationMutation marks an operation as a mutation
	OperationMutation OperationType = "mutation"
	// OperationQuery marks an operation as a query
	OperationQuery OperationType = "query"
)

// GraphQLRequest represents the graphql request body
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// Options defines a GraphQLRequest with a type, so the middlewares know what to do
type Options struct {
	GraphQLRequest
	Type OperationType `json:"type"`
}

var errNoConfigFound = errors.New("grapghql: no configuration found")

// GetOptions extracts the Options config from the backend's extra config
func GetOptions(cfg config.ExtraConfig) (*Options, error) {
	tmp, ok := cfg[Namespace]
	if !ok {
		return nil, errNoConfigFound
	}

	b, err := json.Marshal(tmp)
	if err != nil {
		return nil, err
	}

	var opt Options
	if err := json.Unmarshal(b, &opt); err != nil {
		return nil, err
	}

	return &opt, nil
}

// New resturns a new Extractor, ready to be use on a middleware
func New(opt Options) Extractor {
	replacements := [][2]string{}
	for k, v := range opt.Variables {
		val, ok := v.(string)
		if !ok {
			continue
		}
		if val[0] == '{' && val[len(val)-1] == '}' {
			replacements = append(replacements, [2]string{k, strings.Title(val[1:2]) + val[2:len(val)-1]})
		}
	}

	if len(replacements) == 0 {
		b, _ := json.Marshal(opt.GraphQLRequest)

		return Extractor{
			cfg: opt,
			paramExtractor: func(_ map[string]string) ([]byte, error) {
				return b, nil
			},
		}
	}

	return Extractor{
		cfg: opt,
		paramExtractor: func(params map[string]string) ([]byte, error) {
			val := GraphQLRequest{
				Query:     opt.Query,
				Variables: map[string]interface{}{},
			}
			for k, v := range opt.Variables {
				val.Variables[k] = v
			}
			for _, vs := range replacements {
				val.Variables[vs[0]] = params[vs[1]]
			}
			return json.Marshal(val)
		},
	}
}

// Extractor exposes two extractor factories: one for the params (query) and one
// for the request body (mutator)
type Extractor struct {
	cfg            Options
	paramExtractor func(map[string]string) ([]byte, error)
}

// BodyExtractor returns a graphql request with the given query and the default variables
// overiden by the request body
func (e Extractor) BodyExtractor(r io.Reader) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return b, err
	}
	vars := map[string]interface{}{}

	if err := json.Unmarshal(b, &vars); err != nil {
		return b, err
	}

	for k, v := range e.cfg.Variables {
		if _, ok := vars[k]; ok {
			continue
		}
		vars[k] = v
	}

	request := GraphQLRequest{
		Query:     e.cfg.Query,
		Variables: vars,
	}

	return json.Marshal(request)
}

// ParamExtractor returns a grapql request generator for the given query and the default
// variables overiden by the request params
func (e Extractor) ParamExtractor(params map[string]string) ([]byte, error) {
	return e.paramExtractor(params)
}
