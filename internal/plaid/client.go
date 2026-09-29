package plaid

import (
	"os"
	"sync"

	plaidgo "github.com/plaid/plaid-go/v44/plaid"
)

var (
	client *plaidgo.APIClient
	once   sync.Once
)

/**************************************************************************************
* Client returns a configured singleton Plaid API client. Credentials come from the
* PLAID_CLIENT_ID / PLAID_SECRET env vars and are sent as default headers on every
* request. PLAID_ENV=sandbox (default) or production selects the API environment.
*
* status: ✅
**************************************************************************************/
func Client() *plaidgo.APIClient {
	once.Do(func() {
		configuration := plaidgo.NewConfiguration()
		configuration.AddDefaultHeader("PLAID-CLIENT-ID", os.Getenv("PLAID_CLIENT_ID"))
		configuration.AddDefaultHeader("PLAID-SECRET", os.Getenv("PLAID_SECRET"))

		if os.Getenv("PLAID_ENV") == "production" {
			configuration.UseEnvironment(plaidgo.Production)
		} else {
			configuration.UseEnvironment(plaidgo.Sandbox)
		}

		client = plaidgo.NewAPIClient(configuration)
	})

	return client
}
