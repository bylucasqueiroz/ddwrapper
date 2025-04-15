package ddwrapper

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awsdd "gopkg.in/DataDog/dd-trace-go.v1/contrib/aws/aws-sdk-go-v2/aws"
)

// WrapAWSConfig injects the Datadog middleware into aws.Config
func (w *DDWrapper) WrapAWSConfig(cfg *aws.Config) {
	if !w.config.Enabled {
		return
	}

	// Applies the Datadog wrapper to instrument AWS calls
	awsdd.AppendMiddleware(cfg)
}
