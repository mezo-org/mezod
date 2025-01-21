package cli

import flag "github.com/spf13/pflag"

const (
	FlagRPCURL          = "rpc-url"          // URL of the RPC provider
	FlagMoniker         = "moniker"          // validator's name
	FlagIdentity        = "identity"         // validator's optional identity signature (ex. UPort or Keybase)
	FlagWebsite         = "website"          // validator's optional website
	FlagSecurityContact = "security-contact" // validator's optional security contact email
	FlagDetails         = "details"          // validator's optional details
)

func NewFlagSetSubmitApplication() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagRPCURL, "", "URL of the RPC provider")
	fs.String(FlagIdentity, "", "validator's optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "validator's optional website")
	fs.String(FlagSecurityContact, "", "validator's optional security contact email")
	fs.String(FlagDetails, "", "validator's optional details")

	return fs
}
