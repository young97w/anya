package anya

import "regexp"

type nodeType uint

const (
	nodeStatic nodeType = iota
	nodeStar
	nodeParam
	nodeRegex
)

type node struct {
	path     string
	handler  HandleFunc
	children map[string]*node
	//* match any
	starChild *node
	//match parameter
	// 正则节点 :id(正则表达式)
	paramChild *node
	// 参数key 名称
	param string
	//use regex to match
	regChild *node
	// 放编译好的正则表达式
	regExp *regexp.Regexp

	typ nodeType
	//store middleware
	mdls []Middleware
}
