package imap

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/s-newman/scorestack/dynamicbeat/checks/schema"
)

// The Definition configures the behavior of the imap check
// it implements the "check" interface
type Definition struct {
	Config    schema.CheckConfig // generic metadata about the check
	Host      string             // (required) IP or hostname for the imap server
	Username  string             // (required) Username for the imap server
	Password  string             // (required) Password for the user of the imap server
	Encrypted bool               // (optional, default=false) Whether or not to use TLS (IMAPS)
	Port      string             // (optional, default=143) Port for the imap server
}

// Run a single instance of the check
// We are only supporting the listing of mailboxes as a check currently
func (d *Definition) Run(ctx context.Context, result schema.CheckResult) schema.CheckResult {

	// Set up result
	result := schema.CheckResult{
		Timestamp:   time.Now(),
		ID:          d.Config.ID,
		Name:        d.Config.Name,
		Group:       d.Config.Group,
		ScoreWeight: d.Config.ScoreWeight,
		CheckType:   "imap",
	}

	// Create a dialer so we can set timeouts
	dialer := net.Dialer{
		Timeout: 20 * time.Second,
	}

	// Defining these allow the if/else block below
	var c *client.Client
	var err error

	// Connect to server with TLS or not
	if d.Encrypted {
		c, err = client.DialWithDialerTLS(&dialer, fmt.Sprintf("%s:%s", d.Host, d.Port), &tls.Config{})
	} else {
		c, err = client.DialWithDialer(&dialer, fmt.Sprintf("%s:%s", d.Host, d.Port))
	}
	if err != nil {
		result.Message = fmt.Sprintf("Connecting to server %s failed : %s", d.Host, err)
		return result
	}
	defer func() {
		if closeErr := c.Logout(); closeErr != nil {
			// logp.Warn("failed to close imap connection: %s", closeErr.Error())
		}
	}()

	// Set timeout for commands
	c.Timeout = 5 * time.Second

	// Login
	err = c.Login(d.Username, d.Password)
	if err != nil {
		result.Message = fmt.Sprintf("Login with user %s failed : %s", d.Username, err)
		return result
	}

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	err = c.List("", "*", mailboxes)
	if err != nil {
		result.Message = fmt.Sprintf("Listing mailboxes failed : %s", err)
		return result
	}

	// If we make it here the check passes
	result.Passed = true
	return result

}

// Init the check using a known ID and name. The rest of the check fields will
// be filled in by parsing a JSON string representing the check definition.
func (d *Definition) Init(config schema.CheckConfig, def []byte) error {

	// Set optional values
	d.Port = "143"

	// Unpack JSON definition
	err := json.Unmarshal(def, &d)
	if err != nil {
		return err
	}

	// Set generic values
	d.Config = config

	// Check for missing fields
	missingFields := make([]string, 0)
	if d.Host == "" {
		missingFields = append(missingFields, "Host")
	}

	if d.Username == "" {
		missingFields = append(missingFields, "Username")
	}

	if d.Password == "" {
		missingFields = append(missingFields, "Password")
	}

	// Error only the first missing field, if there are any
	if len(missingFields) > 0 {
		return schema.ValidationError{
			ID:    d.Config.ID,
			Type:  "imap",
			Field: missingFields[0],
		}
	}
	return nil
}

// GetConfig returns the current CheckConfig struct this check has been
// configured with.
func (d *Definition) GetConfig() schema.CheckConfig {
	return d.Config
}
