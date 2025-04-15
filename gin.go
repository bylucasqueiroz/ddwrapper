package ddwrapper

import (
	"github.com/gin-gonic/gin"
	gindd "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

// WrapGin adds the Datadog middleware to Gin
func (w *DDWrapper) WrapGin(r *gin.Engine) {
	if !w.config.Enabled {
		return
	}
	r.Use(gindd.Middleware(w.config.ServiceName))
}
