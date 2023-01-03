package error_handler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

// StdErrorHandler Basic error handler to standardise error response.
// Note this is used for errors which should not crash the app ie non fatal
func StdErrorHandler(cause string, err error) {
	log.Error(fmt.Sprintf("Error: exporter failed due to: %s \n", cause), err)
}
