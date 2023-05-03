package anya

type Middleware func(handleFunc HandleFunc) HandleFunc
