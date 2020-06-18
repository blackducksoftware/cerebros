package stress_testing

import (
	"github.com/blackducksoftware/cerebros/go/pkg/polaris/api"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func RunLoginsForUsers(client *api.Client, userCount int) error {
	pageSize := 10
	pages := (userCount / pageSize) + 1
	for i := 0; i < pages; i++ {
		users, err := client.GetUsers(i*pageSize, pageSize)
		if err != nil {
			return errors.WithMessagef(err, "unable to get users at offset %d, limit %d for logins", i*pageSize, pageSize)
		}
		log.Debugf("got user page with meta %+v", users.Meta)
		for _, user := range users.Data {
			log.Debugf("logging in as %s", user.Attributes.Email)
			loginClient := api.NewClient(client.URL, user.Attributes.Email, "synopsys123")
			err = loginClient.Authenticate()
			if err != nil {
				return errors.WithMessagef(err, "unable to login as user %s", user.Attributes.Email)
			}
			log.Debugf("logged in as %s", user.Attributes.Email)
		}
	}
	return nil
}
