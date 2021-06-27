package treemux

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

type handlerMap struct {
	get     HandlerFunc
	post    HandlerFunc
	put     HandlerFunc
	delete  HandlerFunc
	head    HandlerFunc
	options HandlerFunc
	patch   HandlerFunc

	// If true, the head handler was set implicitly, so let it also be set explicitly.
	implicitHead bool

	m map[string]HandlerFunc
}

func newHandlerMap() *handlerMap {
	return new(handlerMap)
}

func (h *handlerMap) String() string {
	var ss []string
	if h.get != nil {
		ss = append(ss, http.MethodGet)
	}
	if h.post != nil {
		ss = append(ss, http.MethodPost)
	}
	if h.put != nil {
		ss = append(ss, http.MethodPut)
	}
	if h.delete != nil {
		ss = append(ss, http.MethodDelete)
	}
	if h.head != nil {
		ss = append(ss, http.MethodHead)
	}
	if h.options != nil {
		ss = append(ss, http.MethodOptions)
	}
	if h.patch != nil {
		ss = append(ss, http.MethodPatch)
	}
	for k := range h.m {
		ss = append(ss, k)
	}
	sort.Strings(ss)
	return "[" + strings.Join(ss, " ") + "]"
}

func (h *handlerMap) Get(name string) HandlerFunc {
	switch name {
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
		return h.m[name]
	}
}

func (h *handlerMap) Set(name string, handler HandlerFunc) {
	switch name {
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
		if h.m == nil {
			h.m = make(map[string]HandlerFunc)
		}
		h.m[name] = handler
	}
}

type staticNode struct {
	*node
	firstChar byte
}

type node struct {
	route    string
	path     string
	priority int

	// The list of static children to check.
	staticChildren []staticNode

	// If none of the above match, check the wildcard children
	wildcardChild *node

	// If none of the above match, then we use the catch-all, if applicable.
	catchAllChild *node

	// Data for the node is below.

	// If this node is the end of the URL, then call the handler, if applicable.
	handlerMap *handlerMap

	// The names of the parameters to apply.
	leafWildcardNames []string

	addSlash   bool
	isCatchAll bool
}

func (n *node) paramName(i int) string {
	return n.leafWildcardNames[len(n.leafWildcardNames)-1-i]
}

func (n *node) sortStaticChildren(i int) {
	for i > 0 && n.staticChildren[i].priority > n.staticChildren[i-1].priority {
		n.staticChildren[i], n.staticChildren[i-1] = n.staticChildren[i-1], n.staticChildren[i]
		i -= 1
	}
}

func (n *node) setHandler(verb string, handler HandlerFunc, implicitHead bool) {
	if n.handlerMap == nil {
		n.handlerMap = newHandlerMap()
	}
	if h := n.handlerMap.Get(verb); h != nil &&
		(verb != http.MethodHead || !n.handlerMap.implicitHead) {
		panic(fmt.Sprintf("%s already handles %s", n.path, verb))
	}
	n.handlerMap.Set(verb, handler)
	if verb == http.MethodHead {
		n.handlerMap.implicitHead = implicitHead
	}
}

func (n *node) addPath(path string, wildcards []string, inStaticToken bool) *node {
	if path == "" { // leaf
		if wildcards == nil {
			return n
		}

		if n.leafWildcardNames == nil {
			// No wildcards yet, so just add the existing set.
			n.leafWildcardNames = wildcards
			return n
		}

		// Make sure the current wildcards are the same as the old ones.
		// If not then we have an ambiguous path.
		if len(n.leafWildcardNames) != len(wildcards) {
			// This should never happen.
			panic("Reached leaf node with differing wildcard array length. Please report this as a bug.")
		}

		for i := 0; i < len(wildcards); i++ {
			if n.leafWildcardNames[i] != wildcards[i] {
				panic(fmt.Sprintf("Wildcards %v are ambiguous with wildcards %v",
					n.leafWildcardNames, wildcards))
			}
		}

		return n
	}

	c := path[0]
	nextSlash := strings.IndexByte(path, '/')
	var thisToken string
	var tokenEnd int

	if c == '/' {
		// Done processing the previous token, so reset inStaticToken to false.
		thisToken = "/"
		tokenEnd = 1
	} else if nextSlash == -1 {
		thisToken = path
		tokenEnd = len(path)
	} else {
		thisToken = path[0:nextSlash]
		tokenEnd = nextSlash
	}
	remainingPath := path[tokenEnd:]

	if c == '*' && !inStaticToken {
		// Token starts with a *, so it's a catch-all
		thisToken = thisToken[1:]
		if n.catchAllChild == nil {
			n.catchAllChild = &node{path: thisToken, isCatchAll: true}
		}

		if path[1:] != n.catchAllChild.path {
			panic(fmt.Sprintf("Catch-all name in %s doesn't match %s. You probably tried to define overlapping catchalls",
				path, n.catchAllChild.path))
		}

		if nextSlash != -1 {
			panic("/ after catch-all found in " + path)
		}

		if wildcards == nil {
			wildcards = []string{thisToken}
		} else {
			wildcards = append(wildcards, thisToken)
		}
		n.catchAllChild.leafWildcardNames = wildcards

		return n.catchAllChild
	}

	if c == ':' && !inStaticToken {
		// Token starts with a :
		thisToken = thisToken[1:]

		if wildcards == nil {
			wildcards = []string{thisToken}
		} else {
			wildcards = append(wildcards, thisToken)
		}

		if n.wildcardChild == nil {
			n.wildcardChild = &node{path: "wildcard"}
		}

		return n.wildcardChild.addPath(remainingPath, wildcards, false)
	}

	// if strings.ContainsAny(thisToken, ":*") {
	// 	panic("* or : in middle of path component " + path)
	// }

	var unescaped bool

	if len(thisToken) >= 2 && !inStaticToken {
		if thisToken[0] == '\\' && (thisToken[1] == '*' || thisToken[1] == ':' || thisToken[1] == '\\') {
			// The token starts with a character escaped by a backslash. Drop the backslash.
			c = thisToken[1]
			thisToken = thisToken[1:]
			unescaped = true
		}
	}

	// Set inStaticToken to ensure that the rest of this token is not mistaken
	// for a wildcard if a prefix split occurs at a '*' or ':'.
	inStaticToken = (c != '/')

	// Do we have an existing node that starts with the same letter?
	for i, staticChildren := range n.staticChildren {
		if staticChildren.firstChar == c {
			// Yes. Split it based on the common prefix of the existing
			// node and the new one.
			child, prefixSplit := n.splitCommonPrefix(staticChildren.node, thisToken)
			n.staticChildren[i].node = child

			child.priority++
			n.sortStaticChildren(i)

			if unescaped {
				// Account for the removed backslash.
				prefixSplit++
			}
			return child.addPath(path[prefixSplit:], wildcards, inStaticToken)
		}
	}

	// No existing node starting with this letter, so create it.
	child := &node{path: thisToken}

	if n.staticChildren == nil {
		n.staticChildren = []staticNode{{firstChar: c, node: child}}
	} else {
		n.staticChildren = append(n.staticChildren, staticNode{firstChar: c, node: child})
	}
	return child.addPath(remainingPath, wildcards, inStaticToken)
}

func (n *node) splitCommonPrefix(childNode *node, path string) (*node, int) {
	if strings.HasPrefix(path, childNode.path) {
		// No split needs to be done. Rather, the new path shares the entire
		// prefix with the existing node, so the new node is just a child of
		// the existing one. Or the new path is the same as the existing path,
		// which means that we just move on to the next token. Either way,
		// this return accomplishes that
		return childNode, len(childNode.path)
	}

	var i int
	// Find the length of the common prefix of the child node and the new path.
	for i = range childNode.path {
		if i == len(path) {
			break
		}
		if path[i] != childNode.path[i] {
			break
		}
	}

	commonPrefix := path[:i]
	childNode.path = childNode.path[i:]

	// Create a new intermediary node in the place of the existing node, with
	// the existing node as a child.
	newNode := &node{
		path:     commonPrefix,
		priority: childNode.priority,
		staticChildren: []staticNode{{
			firstChar: childNode.path[0],
			node:      childNode,
		}},
	}

	return newNode, i
}

func (n *node) search(method, path string) (*node, HandlerFunc, []Param) {
	// if test != nil {
	// 	test.Logf("Searching for %s in %s", path, n.dumpTree("", ""))
	// }

	if path == "" {
		if n.handlerMap == nil {
			return nil, nil, nil
		}
		return n, n.handlerMap.Get(method), nil
	}

	var lastNode *node

	// First see if this matches a static token.
	firstChar := path[0]
	for _, child := range n.staticChildren {
		if child.firstChar == firstChar {
			if strings.HasPrefix(path, child.path) {
				nextPath := path[len(child.path):]
				node, handler, params := child.search(method, nextPath)
				if handler != nil {
					return node, handler, params
				}
				lastNode = node
			}
			break
		}
	}

	if n.wildcardChild != nil {
		// Didn't find a static token, so check for a wildcard.
		nextSlash := strings.IndexByte(path, '/')
		if nextSlash < 0 {
			nextSlash = len(path)
		}

		thisToken := path[:nextSlash]
		nextToken := path[nextSlash:]

		if len(thisToken) > 0 { // Don't match on empty tokens.
			node, handler, params := n.wildcardChild.search(method, nextToken)
			if handler != nil || (lastNode == nil && node != nil) {
				unescaped, err := url.PathUnescape(thisToken)
				if err != nil {
					unescaped = thisToken
				}

				if params == nil {
					params = make([]Param, 0, len(node.leafWildcardNames))
				}
				params = append(params, Param{
					Name:  node.paramName(len(params)),
					Value: unescaped,
				})

				if handler != nil {
					return node, handler, params
				}

				// Didn't actually find a handler here, so remember that we
				// found a node but also see if we can fall through to the
				// catchall.
				lastNode = node
			}
		}
	}

	if n.catchAllChild != nil {
		// Hit the catchall, so just assign the whole remaining path if it
		// has a matching handler.
		handler := n.catchAllChild.handlerMap.Get(method)
		// Found a handler, or we found a catchall node without a handler.
		// Either way, return it since there's nothing left to check after this.
		if handler != nil || lastNode == nil {
			unescaped, err := url.PathUnescape(path)
			if err != nil {
				unescaped = path
			}

			return n.catchAllChild, handler, []Param{{
				Name:  n.catchAllChild.paramName(0),
				Value: unescaped,
			}}
		}
	}

	return lastNode, nil, nil
}

func (n *node) dumpTree(prefix, nodeType string) string {
	line := fmt.Sprintf("%s %02d %s%s [%d] %v wildcards %v\n", prefix, n.priority, nodeType, n.path,
		len(n.staticChildren), n.handlerMap, n.leafWildcardNames)
	prefix += "  "
	for _, node := range n.staticChildren {
		line += node.dumpTree(prefix, "")
	}
	if n.wildcardChild != nil {
		line += n.wildcardChild.dumpTree(prefix, ":")
	}
	if n.catchAllChild != nil {
		line += n.catchAllChild.dumpTree(prefix, "*")
	}
	return line
}
