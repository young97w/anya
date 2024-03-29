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
			return errInvalidPath(path)
		}
		var err error
		root, err = r.findOrBuild(root, seg)
		if err != nil {
			return err
		}
	}

	if root.handler != nil {
		return errPathRegistered(root.path)
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

func (r *router) findRoute(method, path string) (*nodeInfo, bool) {
	info := &nodeInfo{}
	//special case
	var root *node
	if path == "/" {
		var ok bool
		root, ok = r.m[method]
		if !ok {
			return info, false
		}

		info.n = root
		return info, false
	}

	if path[0] != '/' {
		return info, false
	}

	//start find root
	root, _ = r.m[method]
	for _, seg := range strings.Split(strings.Trim(path, "/"), "/") {
		if seg == "" {
			return info, false
		}

		child, ok, isParam, param := r.findChild(root, seg)
		if !ok {
			//如果是上一个是通配符匹配，回到通配符
			if root.typ == nodeStar {
				return info, true
			}
			return info, false
		}

		// 添加参数
		if isParam {
			if info.params == nil {
				info.params = make(map[string]string, 4)
			}
			info.params[child.param] = param
		}

		root = child
		info.n = root
	}
	return info, true
}

//第二个bool表示返回节点是否有值，第三个为节点是否带参数，第四个为参数
func (r *router) findChild(root *node, seg string) (*node, bool, bool, string) {
	var res *node
	var ok bool
	//static node
	if root.children != nil {
		res, ok = root.children[seg]
		if ok {
			return res, ok, false, ""
		}
	}

	// 特殊节点（三选一）
	switch {
	case root.starChild != nil:
		return root.starChild, true, false, ""
	case root.paramChild != nil:
		return root.paramChild, true, true, seg
	case root.regChild != nil:
		param := root.regChild.regExp.FindString(seg)
		return root.regChild, true, true, param
	}

	return nil, false, false, ""
}

func (r *router) mergeMdls() {
	for _, n := range r.m {
		mergeMdls(n)
	}
}

func mergeMdls(n *node) {
	// find self *
	// find child *
	// child mdls = self + self * + child *
	arr := []*node{n}
	for len(arr) > 0 {
		cur := arr[0]
		arr = arr[1:]
		//find self *
		gnrMdls := cur.mdls
		if cur.starChild != nil {
			gnrMdls = append(cur.mdls, cur.starChild.mdls...)
			cur.starChild.mdls = append(cur.mdls, cur.starChild.mdls...)
			arr = append(arr, cur.starChild)
		} else if cur.paramChild != nil {
			cur.paramChild.mdls = append(cur.mdls, cur.paramChild.mdls...)
			arr = append(arr, cur.paramChild)
		} else if cur.regChild != nil {
			cur.regChild.mdls = append(cur.mdls, cur.regChild.mdls...)
			arr = append(arr, cur.regChild)
		}

		//find child
		for _, child := range cur.children {
			child.mdls = append(gnrMdls, child.mdls...)
			arr = append(arr, child)
		}
	}
}
