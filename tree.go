package httptreemux

import (
	"fmt"
	"net/url"
	"strings"
)

type node struct {
	path string

	priority int

	// The list of static children to check.
	staticIndices []byte
	staticChild   []*node

	// If none of the above match, check the list of wildcard children
	wildcardChild []*node

	// If none of the above match, then we use the catch-all, if applicable.
	catchAllChild *node

	// Data for the node is below.

	addSlash   bool
	isCatchAll bool
	// If this node is the end of the URL, then call the handler, if applicable.
	leafHandler map[string]HandlerFunc
}

func (n *node) sortStaticChild(i int) {
	for i > 0 && n.staticChild[i].priority > n.staticChild[i-1].priority {
		n.staticChild[i], n.staticChild[i-1] = n.staticChild[i-1], n.staticChild[i]
		n.staticIndices[i], n.staticIndices[i-1] = n.staticIndices[i-1], n.staticIndices[i]
		i -= 1
	}
}

func (n *node) sortWildcardChild(i int) {
	for i > 0 && n.wildcardChild[i].priority > n.wildcardChild[i-1].priority {
		n.wildcardChild[i], n.wildcardChild[i-1] = n.wildcardChild[i-1], n.wildcardChild[i]
		i -= 1
	}
}

func (n *node) setHandler(verb string, handler HandlerFunc) {
	if n.leafHandler == nil {
		n.leafHandler = make(map[string]HandlerFunc)
	}
	_, ok := n.leafHandler[verb]
	if ok {
		panic(fmt.Sprintf("%s already handles %s", n.path, verb))
	}
	n.leafHandler[verb] = handler
}

func (n *node) addPath(path string) *node {
	leaf := len(path) == 0
	if leaf {
		return n
	}

	c := path[0]
	nextSlash := strings.Index(path, "/")
	var thisToken string
	var tokenEnd int

	if c == '/' {
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

	if c == '*' {
		// Token starts with a *, so it's a catch-all
		thisToken = thisToken[1:]
		if n.catchAllChild == nil {
			n.catchAllChild = &node{path: thisToken, isCatchAll: true}
		}

		if path[1:] != n.catchAllChild.path {
			panic(fmt.Sprintf("Catch-all name in %s doesn't match %s",
				path, n.catchAllChild.path))
		}

		if nextSlash != -1 {
			panic("/ after catch-all found in " + path)
		}

		return n.catchAllChild
	} else if c == ':' {
		// Token starts with a :
		thisToken = thisToken[1:]
		var child *node
		for i, childNode := range n.wildcardChild {
			// Find a wildcard node with the same name as this one.
			if childNode.path == thisToken {
				child = childNode
				child.priority++
				n.sortWildcardChild(i)
				break
			}
		}

		if child == nil {
			child = &node{path: thisToken}
			if n.wildcardChild == nil {
				n.wildcardChild = []*node{child}
			} else {
				n.wildcardChild = append(n.wildcardChild, child)
			}
		}

		return child.addPath(remainingPath)

	} else {
		if strings.ContainsAny(thisToken, ":*") {
			panic("* or : in middle of path component " + path)
		}

		// Do we have an existing node that starts with the same letter?
		for i, index := range n.staticIndices {
			if c == index {
				// Yes. Split it based on the common prefix of the existing
				// node and the new one.
				child, prefixSplit := n.splitCommonPrefix(i, thisToken)
				child.priority++
				n.sortStaticChild(i)
				return child.addPath(path[prefixSplit:])
			}
		}

		// No existing node starting with this letter, so create it.
		child := &node{path: thisToken}

		if n.staticIndices == nil {
			n.staticIndices = []byte{c}
			n.staticChild = []*node{child}
		} else {
			n.staticIndices = append(n.staticIndices, c)
			n.staticChild = append(n.staticChild, child)
		}
		return child.addPath(remainingPath)
	}
}

func (n *node) splitCommonPrefix(existingNodeIndex int, path string) (*node, int) {
	childNode := n.staticChild[existingNodeIndex]

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

	commonPrefix := path[0:i]
	childNode.path = childNode.path[i:]

	// Create a new intermediary node in the place of the existing node, with
	// the existing node as a child.
	newNode := &node{
		path:     commonPrefix,
		priority: childNode.priority,
		// Index is the first letter of the non-common part of the path.
		staticIndices: []byte{childNode.path[0]},
		staticChild:   []*node{childNode},
	}
	n.staticChild[existingNodeIndex] = newNode

	return newNode, i
}

func (n *node) search(path string, params *map[string]string) (found *node) {
	// if test != nil {
	// 	test.Logf("Searching for %s in %s", path, n.dumpTree("", ""))
	// }
	pathLen := len(path)
	if pathLen == 0 {
		if len(n.leafHandler) == 0 {
			return nil
		} else {
			return n
		}

	}

	// First see if this matches a static token.
	firstChar := path[0]
	for i, staticIndex := range n.staticIndices {
		if staticIndex == firstChar {
			child := n.staticChild[i]
			childPathLen := len(child.path)
			if pathLen >= childPathLen && child.path == path[:childPathLen] {
				nextPath := path[childPathLen:]
				found = child.search(nextPath, params)
			}
			break
		}
	}

	if found != nil {
		return found
	}

	if len(n.wildcardChild) != 0 {
		// Didn't find a static token, so check the wildcards.
		nextSlash := 0
		for nextSlash < pathLen && path[nextSlash] != '/' {
			nextSlash++
		}

		thisToken := path[0:nextSlash]
		nextToken := path[nextSlash:]

		if len(thisToken) > 0 { // Don't match on empty tokens.
			for _, child := range n.wildcardChild {
				found = child.search(nextToken, params)
				if found != nil {
					unescaped, err := url.QueryUnescape(thisToken)
					if err != nil {
						unescaped = thisToken
					}

					if *params == nil {
						*params = map[string]string{child.path: unescaped}
					} else {
						(*params)[child.path] = unescaped
					}

					return
				}
			}
		}
	}

	catchAllChild := n.catchAllChild
	if catchAllChild != nil {
		// Hit the catchall, so just assign the whole remaining path.
		unescaped, err := url.QueryUnescape(path)
		if err != nil {
			unescaped = path
		}

		if *params == nil {
			*params = map[string]string{catchAllChild.path: unescaped}
		} else {
			(*params)[catchAllChild.path] = unescaped
		}
		return catchAllChild
	}

	return nil
}

func (n *node) dumpTree(prefix, nodeType string) string {
	line := fmt.Sprintf("%s %02d %s%s [%d] %v\n", prefix, n.priority, nodeType, n.path,
		len(n.staticChild)+len(n.wildcardChild), n.leafHandler)
	prefix += "  "
	for _, node := range n.staticChild {
		line += node.dumpTree(prefix, "")
	}
	for _, node := range n.wildcardChild {
		line += node.dumpTree(prefix, ":")
	}
	if n.catchAllChild != nil {
		line += n.catchAllChild.dumpTree(prefix, "*")
	}
	return line
}
