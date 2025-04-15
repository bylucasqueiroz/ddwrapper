package ddwrapper

import (
	"net/http"

	httpdd "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

// WrapHTTPHandler adds tracing to an http.Handler
func (w *DDWrapper) WrapHTTPHandler(handler http.Handler) http.Handler {
	if !w.config.Enabled {
		return handler
	}
	// Uses a fixed resource name for compatibility with older versions
	return httpdd.WrapHandler(handler, w.config.ServiceName, "http.request")
}

// WrapHTTPClient instruments an http.Client
func (w *DDWrapper) WrapHTTPClient(client *http.Client) *http.Client {
	if !w.config.Enabled {
		return client
	}
	return httpdd.WrapClient(client)
}
