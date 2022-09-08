package dashboard

import (
	"time"

	"github.com/iotaledger/hive.go/core/app"
)

const (
	maxDashboardAuthUsernameSize = 25
)

// ParametersDashboard contains the definition of the parameters used by WarpSync.
type ParametersDashboard struct {
	// BindAddress defines the bind address on which the dashboard can be accessed from
	BindAddress string `default:"localhost:8081" usage:"the bind address on which the dashboard can be accessed from"`
	// DeveloperMode defines whether to run the dashboard in dev mode
	DeveloperMode bool `default:"false" usage:"whether to run the dashboard in dev mode"`
	// DeveloperModeURL defines the URL to use for dev mode
	DeveloperModeURL string `name:"developerModeURL" default:"http://127.0.0.1:9090" usage:"the URL to use for dev mode"`

	Auth struct {
		// SessionTimeout defines how long the auth session should last before expiring
		SessionTimeout time.Duration `default:"72h" usage:"how long the auth session should last before expiring"`
		// Username defines the auth username
		Username string `default:"admin" usage:"the auth username (max 25 chars)"`
		// PasswordHash defines the auth password+salt as a scrypt hash
		PasswordHash string `default:"0000000000000000000000000000000000000000000000000000000000000000" usage:"the auth password+salt as a scrypt hash"`
		// PasswordSalt defines the auth salt used for hashing the password
		PasswordSalt string `default:"0000000000000000000000000000000000000000000000000000000000000000" usage:"the auth salt used for hashing the password"`
		// IdentityFilePath defines the path to the identity file used for JWT
		IdentityFilePath string `default:"identity.key" usage:"the path to the identity file used for JWT"`
		// Defines the private key used to sign the JWT tokens.
		IdentityPrivateKey string `default:"" usage:"private key used to sign the JWT tokens (optional)"`

		RateLimit struct {
			Enabled     bool          `default:"true" usage:"whether the rate limiting should be enabled"`
			Period      time.Duration `default:"1m" usage:"the period for rate limiting"`
			MaxRequests int           `default:"20" usage:"the maximum number of requests per period"`
			MaxBurst    int           `default:"30" usage:"additional requests allowed in the burst period"`
		}
	}

	// whether the debug logging for requests should be enabled
	DebugRequestLoggerEnabled bool `default:"false" usage:"whether the debug logging for requests should be enabled"`
}

var ParamsDashboard = &ParametersDashboard{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"dashboard": ParamsDashboard,
	},
	Masked: []string{"dashboard.auth.passwordHash", "dashboard.auth.passwordSalt"},
}
