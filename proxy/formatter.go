package proxy

import "strings"

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

// NewEntityFormatter creates an entity formatter with the received params
func NewEntityFormatter(target string, whitelist, blacklist []string, group string, mappings map[string]string) EntityFormatter {
	var propertyFilter propertyFilter
	if len(whitelist) > 0 {
		propertyFilter = newWhitelistingFilter(whitelist)
	} else {
		propertyFilter = newBlacklistingFilter(blacklist)
	}
	sanitizedMappings := make(map[string]string, len(mappings))
	for i, m := range mappings {
		v := strings.Split(m, ".")
		sanitizedMappings[i] = v[0]
	}
	return entityFormatter{
		Target:         target,
		Prefix:         group,
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

func newWhitelistingFilter(whitelist []string) propertyFilter {
	// count num of field names (including repeated ones) to allocate a
	// linear array of strings:
	num_fields := 0
	for _, k := range whitelist {
		num_fields += len(strings.Split(k, "."))
		// TODO: we can "sanitize" and check that there are no empty strings?
	}
	indices := make([]int, len(whitelist))
	fields := make([]string, num_fields)
	f_idx := 0
	for w_idx, k := range whitelist {
		for _, key := range strings.Split(k, ".") {
			fields[f_idx] = key
			f_idx++
		}
		indices[w_idx] = f_idx
	}

	return func(entity *Response) {
		accumulator := make(map[string]interface{}, len(whitelist))
		var p map[string]interface{}
		var c interface{}

		start := 0
		for _, end := range indices {
			p = entity.Data
			ok := true
			d_end := end - 1
			// find 'dicts-path' in input
			for _, field := range fields[start:d_end] {
				if c, ok = p[field]; !ok {
					break
				}
				if p, ok = c.(map[string]interface{}); !ok {
					break
				}
			}
			if ok {
				// find 'value' in input
				if c, ok = p[fields[d_end]]; ok {
					// find 'dicts-path' in accumulator
					d := buildDictPath(accumulator, fields, start, d_end, c)
					// assign 'value' in acuumulator
					d[fields[d_end]] = c
				}
			}
			start = end
		}
		*entity = Response{Data: accumulator, IsComplete: entity.IsComplete}
	}
}

func buildDictPath(accumulator map[string]interface{}, fields []string, f_start int, f_end int, value interface{}) map[string]interface{} {
	var ok bool = true
	var c map[string]interface{}
	var f_idx int
	p := accumulator
	for f_idx = f_start; f_idx < f_end; f_idx++ {
		if c, ok = p[fields[f_idx]].(map[string]interface{}); !ok {
			break
		}
		p = c
	}
	for ; f_idx < f_end; f_idx++ {
		c = make(map[string]interface{})
		p[fields[f_idx]] = c
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
