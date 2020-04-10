package treemux

import "strconv"

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(key string) (string, bool) {
	for _, param := range ps {
		if param.Key == key {
			return param.Value, true
		}
	}
	return "", false
}

func (ps Params) Text(key string) string {
	s, _ := ps.Get(key)
	return s
}

func (ps Params) Uint32(key string) (uint32, error) {
	n, err := strconv.ParseUint(ps.Text(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func (ps Params) Uint64(key string) (uint64, error) {
	return strconv.ParseUint(ps.Text(key), 10, 64)
}

func (ps Params) AsMap() map[string]string {
	if len(ps) == 0 {
		return nil
	}
	m := make(map[string]string, len(ps))
	for _, param := range ps {
		m[param.Key] = param.Value
	}
	return m
}
