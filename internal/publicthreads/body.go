package publicthreads

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	latheruntime "github.com/lathe-cli/lathe/pkg/runtime"
)

// buildCreateBody assembles the create-thread request body the same way the
// generated runtime does for a body with no envelope template: --set/--set-str
// take precedence and merge into one document, otherwise a --file body is used
// verbatim, otherwise nil (an empty thread, which the API allows). The merge
// mirrors Lathe's dotted-path semantics (object fields, array indices, and type
// inference for --set; forced strings for --set-str).
func buildCreateBody(file string, sets, stringSets []string) ([]byte, error) {
	if len(sets) > 0 || len(stringSets) > 0 {
		out := map[string]any{}
		for _, kv := range sets {
			path, value, err := parseSet(kv, "--set")
			if err != nil {
				return nil, err
			}
			if err := setNestedPath(out, path, inferValue(value)); err != nil {
				return nil, err
			}
		}
		for _, kv := range stringSets {
			path, value, err := parseSet(kv, "--set-str")
			if err != nil {
				return nil, err
			}
			if err := setNestedPath(out, path, value); err != nil {
				return nil, err
			}
		}
		return json.Marshal(out)
	}
	if file != "" {
		return latheruntime.ReadBody(file)
	}
	return nil, nil
}

func parseSet(kv, flag string) (string, string, error) {
	eq := strings.Index(kv, "=")
	if eq < 0 {
		return "", "", fmt.Errorf("invalid %s %q (expected key=value)", flag, kv)
	}
	path := kv[:eq]
	if path == "" {
		return "", "", fmt.Errorf("invalid %s %q (empty key)", flag, kv)
	}
	return path, kv[eq+1:], nil
}

type pathSegment struct {
	key string
	idx int // -1 = object field, >=0 = array index within key
}

func parsePath(path string) []pathSegment {
	parts := strings.Split(path, ".")
	segs := make([]pathSegment, 0, len(parts))
	for _, p := range parts {
		if open := strings.Index(p, "["); open >= 0 && strings.HasSuffix(p, "]") {
			key := p[:open]
			if idx, err := strconv.Atoi(p[open+1 : len(p)-1]); err == nil {
				segs = append(segs, pathSegment{key: key, idx: idx})
				continue
			}
		}
		segs = append(segs, pathSegment{key: p, idx: -1})
	}
	return segs
}

func setNestedPath(m map[string]any, path string, v any) error {
	return setNestedSegs(m, parsePath(path), v)
}

func setNestedSegs(m map[string]any, segs []pathSegment, v any) error {
	if len(segs) == 0 {
		return nil
	}
	seg := segs[0]
	rest := segs[1:]

	if seg.idx < 0 {
		if len(rest) == 0 {
			m[seg.key] = v
			return nil
		}
		switch next := m[seg.key].(type) {
		case map[string]any:
			return setNestedSegs(next, rest, v)
		case nil:
			child := map[string]any{}
			m[seg.key] = child
			return setNestedSegs(child, rest, v)
		default:
			return fmt.Errorf("conflicting --set: %s is not an object", seg.key)
		}
	}

	var arr []any
	switch existing := m[seg.key].(type) {
	case []any:
		arr = existing
	case nil:
		arr = []any{}
	default:
		return fmt.Errorf("conflicting --set: %s is not an array", seg.key)
	}
	for len(arr) <= seg.idx {
		arr = append(arr, nil)
	}
	if len(rest) == 0 {
		arr[seg.idx] = v
	} else {
		var child map[string]any
		switch existing := arr[seg.idx].(type) {
		case map[string]any:
			child = existing
		case nil:
			child = map[string]any{}
		default:
			return fmt.Errorf("conflicting --set: %s[%d] is not an object", seg.key, seg.idx)
		}
		if err := setNestedSegs(child, rest, v); err != nil {
			return err
		}
		arr[seg.idx] = child
	}
	m[seg.key] = arr
	return nil
}

func inferValue(s string) any {
	switch s {
	case "true":
		return true
	case "false":
		return false
	case "null":
		return nil
	}
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}
