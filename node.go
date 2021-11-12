package bunrouter

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type node struct {
	route string
	part  string

	params     map[string]int // param name => param position
	handlerMap *handlerMap

	parent     *node
	colon      *node
	isWildcard bool

	nodes []*node
	index struct {
		table   []uint8 // index table for the nodes: firstChar-minChar => node position
		minChar byte    // min char in the table
		maxChar byte    // max char in the table
	}
}

func (n *node) addRoute(route string) (*node, bool) {
	if route == "/" {
		return n, false
	}

	parts, params, slash := splitRoute(route)
	currNode := n

	for _, part := range parts {
		currNode = currNode.addPart(part)
	}

	currNode.route = route
	currNode.params = params
	n.indexNodes()

	return currNode, slash
}

func (n *node) addPart(part string) *node {
	if part == "*" {
		n.isWildcard = true
		return n
	}

	if part == ":" {
		if n.colon == nil {
			n.colon = &node{part: ":"}
		}
		return n.colon
	}

	for childNodeIndex, childNode := range n.nodes {
		if childNode.part[0] != part[0] {
			continue
		}

		// Check for a common prefix.

		for i, c := range []byte(part) {
			if i == len(childNode.part) {
				break
			}
			if c == childNode.part[i] {
				continue
			}

			// Create a node for the common prefix.

			childNode.part = childNode.part[i:]
			newNode := &node{part: part[i:]}

			n.nodes[childNodeIndex] = &node{
				part:  part[:i], // common prefix
				nodes: []*node{childNode, newNode},
			}

			return newNode
		}

		// Parts match completely.

		switch {
		case len(part) > len(childNode.part): // part is bigger
			part = part[len(childNode.part):]
			return childNode.addPart(part)

		case len(part) < len(childNode.part): // part is smaller
			childNode.part = childNode.part[len(part):]
			newNode := &node{part: part}
			newNode.nodes = []*node{childNode}
			n.nodes[childNodeIndex] = newNode
			return newNode

		default:
			return childNode // exact match
		}
	}

	node := &node{part: part}
	n.nodes = append(n.nodes, node)
	return node
}

func (n *node) findRoute(meth, path string) (*node, routeHandler, int) {
	if path == "" {
		return n, n.handler(meth), 0
	}

	var found *node

	if firstChar := path[0]; firstChar >= n.index.minChar && firstChar <= n.index.maxChar {
		if i := n.index.table[firstChar-n.index.minChar]; i != 0 {
			childNode := n.nodes[i-1]

			if childNode.part == path {
				if handler := childNode.handler(meth); handler.fn != nil {
					return childNode, handler, 0
				}
				if childNode.handlerMap != nil {
					found = childNode
				}
			} else {
				partLen := len(childNode.part)
				if strings.HasPrefix(path, childNode.part) {
					node, handler, wildcardLen := childNode.findRoute(meth, path[partLen:])
					if handler.fn != nil {
						return node, handler, wildcardLen
					}
					if node != nil {
						found = node
					}
				}
			}
		}
	}

	if n.colon != nil {
		if i := strings.IndexByte(path, '/'); i > 0 {
			node, handler, wildcardLen := n.colon.findRoute(meth, path[i:])
			if handler.fn != nil {
				return node, handler, wildcardLen
			}
		} else {
			if handler := n.colon.handler(meth); handler.fn != nil {
				return n.colon, handler, 0
			}
			if found == nil && n.colon.handlerMap != nil {
				found = n.colon
			}
		}
	}

	if n.isWildcard && n.handlerMap != nil {
		if handler := n.handler(meth); handler.fn != nil {
			if handler.slash {
				return n, handler, len(path) - 1
			}
			return n, handler, len(path)
		}
		if found == nil {
			found = n
		}
	}

	return found, routeHandler{}, 0
}

func (n *node) indexNodes() {
	if len(n.nodes) > 0 {
		sort.Slice(n.nodes, func(i, j int) bool {
			return n.nodes[i].part[0] < n.nodes[j].part[0]
		})

		n.index.minChar = n.nodes[0].part[0]
		n.index.maxChar = n.nodes[len(n.nodes)-1].part[0]

		// Reset index.
		if size := int(n.index.maxChar - n.index.minChar + 1); len(n.index.table) != size {
			n.index.table = make([]uint8, size)
		} else {
			for i := range n.index.table {
				n.index.table[i] = 0
			}
		}

		// Index nodes by the first char in a part.
		for childNodeIndex, childNode := range n.nodes {
			childNode.parent = n
			childNode.indexNodes()

			firstChar := childNode.part[0] - n.index.minChar
			n.index.table[firstChar] = uint8(childNodeIndex + 1)
		}
	}

	if n.colon != nil {
		n.colon.parent = n
		n.colon.indexNodes()
	}
}

func (n *node) setHandler(verb string, handler routeHandler) {
	if n.handlerMap == nil {
		n.handlerMap = newHandlerMap()
	}
	if h := n.handlerMap.Get(verb); h.fn != nil {
		panic(fmt.Sprintf("%s already handles %s", n.route, verb))
	}
	n.handlerMap.Set(verb, handler)
}

func (n *node) handler(verb string) routeHandler {
	if n.handlerMap == nil {
		return routeHandler{}
	}
	return n.handlerMap.Get(verb)
}

//------------------------------------------------------------------------------

type routeParser struct {
	segments []string
	i        int

	acc   []string
	parts []string
}

func (p *routeParser) valid() bool {
	return p.i < len(p.segments)
}

func (p *routeParser) next() string {
	s := p.segments[p.i]
	p.i++
	return s
}

func (p *routeParser) accumulate(s string) {
	p.acc = append(p.acc, s)
}

func (p *routeParser) finalizePart(withSlash bool) {
	if part := join(p.acc, withSlash); part != "" {
		p.parts = append(p.parts, part)
	}
	p.acc = p.acc[:0]

	if p.valid() {
		p.acc = append(p.acc, "")
	}
}

func join(ss []string, withSlash bool) string {
	if len(ss) == 0 {
		return ""
	}
	s := strings.Join(ss, "/")
	if withSlash {
		return s + "/"
	}
	return s
}

func splitRoute(route string) (_ []string, _ map[string]int, slash bool) {
	if route == "" || route[0] != '/' {
		panic(fmt.Errorf("invalid route: %q", route))
	}

	if route == "/" {
		return []string{}, nil, false
	}
	route = route[1:] // trim "/"

	if strings.HasSuffix(route, "/") {
		slash = true
		route = route[:len(route)-1]
	}

	ss := strings.Split(route, "/")
	if len(ss) == 0 {
		panic(fmt.Errorf("invalid route: %q", route))
	}

	p := routeParser{
		segments: ss,
	}
	var params []string

	for p.valid() {
		segment := p.next()

		if segment == "" {
			p.accumulate("")
			continue
		}

		switch firstChar := segment[0]; firstChar {
		case ':':
			p.finalizePart(true)
			p.parts = append(p.parts, string(firstChar))
			params = append(params, segment[1:])
		case '*':
			p.finalizePart(false)
			slash = len(p.parts) > 0
			p.parts = append(p.parts, string(firstChar))
			params = append(params, segment[1:])
		default:
			p.accumulate(segment)
		}
	}

	p.finalizePart(false)

	if len(params) > 0 {
		return p.parts, paramMap(route, params), slash
	}
	return p.parts, nil, slash
}

func paramMap(route string, params []string) map[string]int {
	m := make(map[string]int, len(params))
	for i, param := range params {
		if param == "" {
			panic(fmt.Errorf("param must have a name: %q", route))
		}
		m[param] = i
	}
	return m
}

//------------------------------------------------------------------------------

type handlerMap struct {
	get        routeHandler
	post       routeHandler
	put        routeHandler
	delete     routeHandler
	head       routeHandler
	options    routeHandler
	patch      routeHandler
	notAllowed routeHandler
}

type routeHandler struct {
	fn    HandlerFunc
	slash bool // redirect with/without a slash
}

func newHandlerMap() *handlerMap {
	return new(handlerMap)
}

func (h *handlerMap) Get(meth string) routeHandler {
	switch meth {
	case http.MethodGet:
		return h.get
	case http.MethodPost:
		return h.post
	case http.MethodPut:
		return h.put
	case http.MethodDelete:
		return h.delete
	case http.MethodHead:
		return h.head
	case http.MethodOptions:
		return h.options
	case http.MethodPatch:
		return h.patch
	default:
		return routeHandler{}
	}
}

func (h *handlerMap) Set(meth string, handler routeHandler) {
	switch meth {
	case http.MethodGet:
		h.get = handler
	case http.MethodPost:
		h.post = handler
	case http.MethodPut:
		h.put = handler
	case http.MethodDelete:
		h.delete = handler
	case http.MethodHead:
		h.head = handler
	case http.MethodOptions:
		h.options = handler
	case http.MethodPatch:
		h.patch = handler
	default:
		panic(fmt.Errorf("unknown HTTP method: %s", meth))
	}
}
