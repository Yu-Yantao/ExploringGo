package middleware

import "net/http"

type Demo2Middleware struct {
}

func NewDemo2Middleware() *Demo2Middleware {
	return &Demo2Middleware{}
}

func (m *Demo2Middleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO generate middleware implement function, delete after code implementation

		// Passthrough to next handler if need
		next(w, r)
	}
}
