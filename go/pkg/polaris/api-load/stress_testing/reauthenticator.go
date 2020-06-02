package stress_testing

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	log "github.com/sirupsen/logrus"
	"time"
)

func Reauthenticator(polarisClient *api.Client, stop <-chan struct{}) {
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(55 * time.Minute):
			}
			for {
				// TODO error-handling?  backoff?
				err := polarisClient.Authenticate()
				recordEvent("reauthentication", err)
				if err == nil {
					log.Infof("successfully reauthenticated")
					break
				}
				log.Errorf("unable to reauthenticate: %+v", err)
			}
		}
	}()
}
