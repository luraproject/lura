// SPDX-License-Identifier: Apache-2.0

package proxy

import (
	"context"
	"fmt"
	"strings"

	"github.com/krakendio/flatmap/tree"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
)

// EntityFormatter formats the response data
type EntityFormatter interface {
	Format(Response) Response
}

// EntityFormatterFunc holds the formatter function
type EntityFormatterFunc func(Response) Response

// Format implements the EntityFormatter interface
func (e EntityFormatterFunc) Format(entity Response) Response { return e(entity) }

type propertyFilter func(*Response)

type entityFormatter struct {
	Target         string
	Prefix         string
	PropertyFilter propertyFilter
	Mapping        map[string]string
}

// NewEntityFormatter creates an entity formatter with the received backend definition
func NewEntityFormatter(remote *config.Backend) EntityFormatter {
	if ef := newFlatmapFormatter(remote.ExtraConfig, remote.Target, remote.Group); ef != nil {
		return ef
	}

	var propertyFilter propertyFilter
	if len(remote.AllowList) > 0 {
		propertyFilter = newAllowlistingFilter(remote.AllowList)
	} else {
		propertyFilter = newDenylistingFilter(remote.DenyList)
	}
	sanitizedMappings := make(map[string]string, len(remote.Mapping))
	for i, m := range remote.Mapping {
		v := strings.Split(m, ".")
		sanitizedMappings[i] = v[0]
	}
	return entityFormatter{
		Target:         remote.Target,
		Prefix:         remote.Group,
		PropertyFilter: propertyFilter,
		Mapping:        sanitizedMappings,
	}
}

// Format implements the EntityFormatter interface
func (e entityFormatter) Format(entity Response) Response {
	if e.Target != "" {
		extractTarget(e.Target, &entity)
	}
	if len(entity.Data) > 0 {
		e.PropertyFilter(&entity)
	}
	if len(entity.Data) > 0 {
		for formerKey, newKey := range e.Mapping {
			if v, ok := entity.Data[formerKey]; ok {
				entity.Data[newKey] = v
				delete(entity.Data, formerKey)
			}
		}
	}
	if e.Prefix != "" {
		entity.Data = map[string]interface{}{e.Prefix: entity.Data}
	}
	return entity
}

func extractTarget(target string, entity *Response) {
	for _, part := range strings.Split(target, ".") {
		if tmp, ok := entity.Data[part]; ok {
			entity.Data, ok = tmp.(map[string]interface{})
			if !ok {
				entity.Data = map[string]interface{}{}
				return
			}
		} else {
			entity.Data = map[string]interface{}{}
			return
		}
	}
}

func AllowlistPrune(wlDict, inDict map[string]interface{}) bool {
	canDelete := true
	var deleteSibling bool
	for k, v := range inDict {
		deleteSibling = true
		if subWl, ok := wlDict[k]; ok {
			if subWlDict, okk := subWl.(map[string]interface{}); okk {
				if subInDict, isDict := v.(map[string]interface{}); isDict && !AllowlistPrune(subWlDict, subInDict) {
					deleteSibling = false
				}
			} else {
				// Allowlist leaf, maintain this branch
				deleteSibling = false
			}
		}
		if deleteSibling {
			delete(inDict, k)
		} else {
			canDelete = false
		}
	}
	return canDelete
}

func newAllowlistingFilter(Allowlist []string) propertyFilter {
	wlDict := make(map[string]interface{})
	for _, k := range Allowlist {
		wlFields := strings.Split(k, ".")
		d := buildDictPath(wlDict, wlFields[:len(wlFields)-1])
		d[wlFields[len(wlFields)-1]] = true
	}

	return func(entity *Response) {
		if AllowlistPrune(wlDict, entity.Data) {
			for k := range entity.Data {
				delete(entity.Data, k)
			}
		}
	}
}

func buildDictPath(accumulator map[string]interface{}, fields []string) map[string]interface{} {
	var ok bool
	var c map[string]interface{}
	var fIdx int
	fEnd := len(fields)
	p := accumulator
	for fIdx = 0; fIdx < fEnd; fIdx++ {
		if c, ok = p[fields[fIdx]].(map[string]interface{}); !ok {
			break
		}
		p = c
	}
	for ; fIdx < fEnd; fIdx++ {
		c = make(map[string]interface{})
		p[fields[fIdx]] = c
		p = c
	}
	return p
}

func buildDenyTree(path []string, tree map[string]interface{}) {
	if len(path) == 0 {
		return
	}
	n := path[0]
	if len(path) == 1 {
		// this is the node to be deleted, so, any other child
		// that is under this node, does not need to be visited:
		// we "delete" any descendant from this node
		tree[n] = nil
		return
	}

	if k, ok := tree[n]; ok {
		if k == nil {
			// all this child should be deleted, so, no matter
			// if the entry says to delete some extra child..
			// everything will be deleted
			return
		}
		childTree, ok := k.(map[string]interface{})
		if !ok {
			// this should never happen if this algorithm is correct
			tree[n] = nil
			return
		}
		buildDenyTree(path[1:], childTree)
		return
	}

	// it the key does not exist, we need to keep building the children,
	// and at this point we know that path is at least len = 2, and that
	// tree[n] does not exist
	childTree := make(map[string]interface{}, 1)
	tree[n] = childTree
	buildDenyTree(path[1:], childTree)
}

func recDelete(ref map[string]interface{}, v interface{}) {
	m, ok := v.(map[string]interface{})
	if !ok || m == nil {
		return
	}

	for rk, rv := range ref {
		dv, dok := m[rk]
		if !dok {
			continue
		}
		if rv == nil {
			delete(m, rk)
			continue
		}
		recDelete(rv.(map[string]interface{}), dv)
	}
}

func newDenylistingFilter(blacklist []string) propertyFilter {
	bl := make(map[string]interface{}, len(blacklist))
	for _, key := range blacklist {
		keys := strings.Split(key, ".")
		buildDenyTree(keys, bl)
	}

	return func(entity *Response) {
		recDelete(bl, entity.Data)
	}
}

const flatmapKey = "flatmap_filter"

type flatmapFormatter struct {
	Target string
	Prefix string
	Ops    []flatmapOp
}

type flatmapOp struct {
	Type string
	Args [][]string
}

// Format implements the EntityFormatter interface
func (e flatmapFormatter) Format(entity Response) Response {
	if e.Target != "" {
		extractTarget(e.Target, &entity)
	}

	e.processOps(&entity)

	if e.Prefix != "" {
		entity.Data = map[string]interface{}{e.Prefix: entity.Data}
	}
	return entity
}

func (e flatmapFormatter) processOps(entity *Response) {
	flatten, err := tree.New(entity.Data)
	if err != nil {
		return
	}
	for _, op := range e.Ops {
		switch op.Type {
		case "move":
			flatten.Move(op.Args[0], op.Args[1])
		case "append":
			flatten.Append(op.Args[0], op.Args[1])
		case "del":
			for _, k := range op.Args {
				flatten.Del(k)
			}
		default:
		}
	}

	entity.Data, _ = flatten.Get([]string{}).(map[string]interface{})
}

func newFlatmapFormatter(cfg config.ExtraConfig, target, group string) *flatmapFormatter {
	if v, ok := cfg[Namespace]; ok {
		if e, ok := v.(map[string]interface{}); ok {
			if vs, ok := e[flatmapKey].([]interface{}); ok {
				if len(vs) == 0 {
					return nil
				}
				var ops []flatmapOp
				for _, v := range vs {
					m, ok := v.(map[string]interface{})
					if !ok {
						continue
					}
					op := flatmapOp{}
					if t, ok := m["type"].(string); ok {
						op.Type = t
					} else {
						continue
					}
					if args, ok := m["args"].([]interface{}); ok {
						op.Args = make([][]string, len(args))
						for k, arg := range args {
							if t, ok := arg.(string); ok {
								op.Args[k] = strings.Split(t, ".")
							}
						}
					}
					ops = append(ops, op)
				}
				if len(ops) == 0 {
					return nil
				}
				return &flatmapFormatter{
					Target: target,
					Prefix: group,
					Ops:    ops,
				}
			}
		}
	}
	return nil
}

// NewFlatmapMiddleware creates a proxy middleware that enables applying flatmap operations to the proxy response
func NewFlatmapMiddleware(logger logging.Logger, cfg *config.EndpointConfig) Middleware {
	formatter := newFlatmapFormatter(cfg.ExtraConfig, "", "")
	return func(next ...Proxy) Proxy {
		if len(next) != 1 {
			panic(ErrTooManyProxies)
		}

		if formatter == nil {
			return next[0]
		}

		logger.Debug(
			fmt.Sprintf(
				"[ENDPOINT: %s][Flatmap] Adding flatmap manipulator with %d operations",
				cfg.Endpoint,
				len(formatter.Ops),
			),
		)

		return func(ctx context.Context, request *Request) (*Response, error) {
			resp, err := next[0](ctx, request)
			if err != nil {
				return resp, err
			}
			r := formatter.Format(*resp)
			return &r, nil
		}
	}
}
