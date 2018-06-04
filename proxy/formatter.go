package proxy

import (
	"strings"

	"github.com/devopsfaith/krakend/config"
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
	var propertyFilter propertyFilter
	if len(remote.Whitelist) > 0 {
		propertyFilter = newWhitelistingFilter(remote.Whitelist)
	} else {
		propertyFilter = newBlacklistingFilter(remote.Blacklist)
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
	if tmp, ok := entity.Data[target]; ok {
		entity.Data, ok = tmp.(map[string]interface{})
		if !ok {
			entity.Data = map[string]interface{}{}
		}
	} else {
		entity.Data = map[string]interface{}{}
	}
}

func whitelistPrune(wlDict map[string]interface{}, inDict map[string]interface{}) bool {
	canDelete := true
	var deleteSibling bool
	for k, v := range inDict {
		deleteSibling = true
		if subWl, ok := wlDict[k]; ok {
			if subWlDict, okk := subWl.(map[string]interface{}); okk {
				if subInDict, isDict := v.(map[string]interface{}); isDict && !whitelistPrune(subWlDict, subInDict) {
					deleteSibling = false
				}
			} else {
				// whitelist leaf, maintain this branch
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

func newWhitelistingFilter(whitelist []string) propertyFilter {
	wlDict := make(map[string]interface{})
	for _, k := range whitelist {
		wlFields := strings.Split(k, ".")
		d := buildDictPath(wlDict, wlFields[:len(wlFields)-1])
		d[wlFields[len(wlFields)-1]] = true
	}

	return func(entity *Response) {
		if whitelistPrune(wlDict, entity.Data) {
			for k := range entity.Data {
				delete(entity.Data, k)
			}
		}
	}
}

func buildDictPath(accumulator map[string]interface{}, fields []string) map[string]interface{} {
	var ok bool = true
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

func newBlacklistingFilter(blacklist []string) propertyFilter {
	bl := make(map[string][]string, len(blacklist))
	for _, key := range blacklist {
		keys := strings.Split(key, ".")
		if len(keys) > 1 {
			if sub, ok := bl[keys[0]]; ok {
				bl[keys[0]] = append(sub, keys[1])
			} else {
				bl[keys[0]] = []string{keys[1]}
			}
		} else {
			bl[keys[0]] = []string{}
		}
	}

	return func(entity *Response) {
		for k, sub := range bl {
			if len(sub) == 0 {
				delete(entity.Data, k)
			} else {
				if tmp := blacklistFilterSub(entity.Data[k], sub); len(tmp) > 0 {
					entity.Data[k] = tmp
				}
			}
		}
	}
}

func blacklistFilterSub(v interface{}, blacklist []string) map[string]interface{} {
	tmp, ok := v.(map[string]interface{})
	if !ok {
		return map[string]interface{}{}
	}
	for _, key := range blacklist {
		delete(tmp, key)
	}
	return tmp
}
