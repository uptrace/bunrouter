package treemux

import "strconv"

type Param struct {
	Name  string
	Value string
}

type Params []Param

func (ps Params) Get(name string) (string, bool) {
	for _, param := range ps {
		if param.Name == name {
			return param.Value, true
		}
	}
	return "", false
}

func (ps Params) Text(name string) string {
	s, _ := ps.Get(name)
	return s
}

func (ps Params) Uint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(ps.Text(name), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint32(n), nil
}

func (ps Params) Uint64(name string) (uint64, error) {
	return strconv.ParseUint(ps.Text(name), 10, 64)
}

func (ps Params) Map() map[string]string {
	if len(ps) == 0 {
		return nil
	}
	m := make(map[string]string, len(ps))
	for _, param := range ps {
		m[param.Name] = param.Value
	}
	return m
}
