package profile

import (
	"cloud.google.com/go/profiler"
	"github.com/goccha/envar"
	"github.com/goccha/log"
	"time"
)

func Activate(serviceName, version string, retry int) error {
	retry++
	var err error
	for i := 0; i < retry; i++ {
		if err = profiler.Start(profiler.Config{
			Service:        serviceName,
			ServiceVersion: version,
			// ProjectID must be set if not running on GCP.
			ProjectID: envar.String("GCP_PROJECT", "GOOGLE_CLOUD_PROJECT"),
		}); err == nil {
			break
		}
		log.Warn("%+v", err)
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	return nil
}
