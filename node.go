package anya

type node struct {
	path     string
	handler  HandleFunc
	children []*node
	//* match any
	starChild *node
	//match parameter
	paramChild *node
	//use regex to match
	regChild *node
	//store middleware
	mdls []Middleware
}
