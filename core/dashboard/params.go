package dashboard

import (
	"time"

	"github.com/iotaledger/hive.go/app"
)

const (
	maxDashboardAuthUsernameSize = 25
)

// ParametersDashboard contains the definition of the parameters used by WarpSync.
type ParametersDashboard struct {
	// BindAddress defines the bind address on which the dashboard can be accessed from
	BindAddress string `default:"localhost:8081" usage:"the bind address on which the dashboard can be accessed from"`
	// Alias is used to set an alias to identify a node
	Alias string `default:"HORNET node" usage:"set an alias to identify a node"`
	// DevMode defines whether to run the dashboard in dev mode
	DevMode bool `name:"dev" default:"false" usage:"whether to run the dashboard in dev mode"`

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
	}
}

var ParamsDashboard = &ParametersDashboard{}

var params = &app.ComponentParams{
	Params: map[string]any{
		"dashboard": ParamsDashboard,
	},
	Masked: []string{"dashboard.auth.passwordHash", "dashboard.auth.passwordSalt"},
}
