package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies a sequence of middleware to a handler.
// Middleware is applied in the order provided, meaning the first middleware
// in the list is the outermost wrapper.
func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	// Apply in reverse so the first middleware is the outermost
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
