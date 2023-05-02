package anya

import (
	"regexp"
	"strings"
)

type router struct {
	m map[string]*node
}

type nodeInfo struct {
	n      *node
	params map[string]string
}

func (r *router) addRoute(method string, path string, handleFunc HandleFunc, mdls ...Middleware) error {
	// 1.check http method tree
	// 2.find tree
	// 3.find parent node
	// 4.add

	//special case
	if path == "/" {
		_, ok := r.m[method]
		if ok {
			return errPathRegistered(path)
		}
		r.m[method] = &node{
			path:    "/",
			handler: handleFunc,
			mdls:    mdls,
		}
		return nil
	}

	if path[0] != '/' {
		return errInvalidPath(path)
	}

	root, ok := r.m[method]
	if !ok {
		root = &node{
			path: "/",
		}
		r.m[method] = root
	}

	segs := strings.Split(strings.Trim(path, "/"), "/")
	for _, seg := range segs {
		if seg == "" {
			panic(errInvalidPath(path))
		}
		var err error
		root, err = r.findOrBuild(root, seg)
		if err != nil {
			return err
		}
	}

	root.handler = handleFunc
	root.mdls = mdls
	return nil
}

func (r *router) findOrBuild(root *node, seg string) (*node, error) {
	//特殊节点： *
	//parameter，以：开头
	//regex
	var res *node
	if seg == "*" {
		//校验剩下的特殊节点是否为空
		if root.paramChild != nil {
			return nil, errNodeConflict("parameter", "*")
		}
		if root.regChild != nil {
			return nil, errNodeConflict("regex", "*")

		}

		if root.starChild != nil {
			return root.starChild, nil
		}

		root.starChild = &node{
			path: seg,
			typ:  nodeStar,
		}
		return root.starChild, nil
	}

	if seg[0] == ':' {
		regSlice := regexp.MustCompile(`^:(\w+)\((.*)\)$`).FindStringSubmatch(seg)
		//regexp.MustCompile(`^:(\w+)\((\w+)\)$`).FindStringSubmatch(seg)
		//校验剩下的特殊节点是否为空
		//普通参数路径
		if len(regSlice) == 0 {
			if root.paramChild != nil {
				return root.paramChild, nil
			}
			//创建前先检验
			if root.starChild != nil {
				return nil, errNodeConflict("*", "parameter")
			}
			if root.regChild != nil {
				return nil, errNodeConflict("regex", "parameter")
			}

			root.paramChild = &node{
				path:  seg,
				param: seg[1:],
				typ:   nodeParam,
			}

			return root.paramChild, nil
		}

		//正则路径, slice 中依次为 原始的 string、param、regexp
		if root.regChild != nil {
			return root.regChild, nil
		}

		//创建前先检验
		if root.starChild != nil {
			return nil, errNodeConflict("*", "regex")
		}
		if root.paramChild != nil {
			return nil, errNodeConflict("parameter", "regex")
		}
		root.regChild = &node{
			path:   seg,
			param:  regSlice[1],
			regExp: regexp.MustCompile(regSlice[2]),
			typ:    nodeRegex,
		}

		return root.regChild, nil
	}

	if root.children == nil {
		root.children = make(map[string]*node, 4)
	}

	res, ok := root.children[seg]
	if ok {
		return res, nil
	}

	res = &node{path: seg, typ: nodeStatic}
	root.children[seg] = res
	return res, nil

}

func (r *router) findRoute(method, path string) (*nodeInfo, error) {
	var info *nodeInfo
	//special case
	var root *node
	if path == "/" {
		var ok bool
		root, ok = r.m[method]
		if !ok {
			return nil, errInvalidPath(path)
		}

		info.n = root
		return info, nil
	}

	if path[0] != '/' {
		return nil, errInvalidPath(path)
	}

	//start find root
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		if seg == "" {
			return nil, errInvalidPath(seg)
		}

		child, ok, isParam, param := r.findChild(root, seg)
		if !ok {
			//如果是上一个是通配符匹配，回到通配符
			if root.typ == nodeStar {
				return info, nil
			}
			return nil, errRouteNotExist(path)
		}

		// 添加参数
		if isParam {
			if info.params == nil {
				info.params = make(map[string]string, 4)
			}
			info.params[root.param] = param
		}

		root = child
		info.n = root
	}
	return nil, errRouteNotExist(path)
}

//第二个bool返回节点是否为nil，第三个为节点是否带参数，第四个为参数
func (r *router) findChild(root *node, seg string) (*node, bool, bool, string) {
	var res *node
	var ok bool
	switch root.typ {
	case nodeStatic:
		res, ok = root.children[seg]
		return res, ok, false, ""
	case nodeStar:
		return root.starChild, root.starChild == nil, false, ""
	case nodeParam:
		return root.paramChild, root.paramChild == nil, true, seg
	case nodeRegex:
		return root.regChild, root.regChild == nil, true, root.regExp.FindString(seg)
	}

	return nil, false, false, ""
}
