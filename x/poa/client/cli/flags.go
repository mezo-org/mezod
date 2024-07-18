package cli

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	flag "github.com/spf13/pflag"
)

const (
	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"
	FlagIP              = "ip"
	FlagP2PPort         = "p2p-port"
	FlagGenValDir       = "genval-dir"
)

func addValidatorDescriptionFlags(fs *flag.FlagSet) *flag.FlagSet {
	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")

	return fs
}

func NewFlagSetSubmitApplication() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs = addValidatorDescriptionFlags(fs)
	return fs
}

func NewFlagSetGenVal(defaultHome, defaultIP, defaultP2PPort string) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(flags.FlagHome, defaultHome, "The application home directory")
	fs.String(
		flags.FlagOutputDocument,
		"",
		"Write the generated validator to the given file instead of the default location",
	)
	fs.String(flags.FlagChainID, "", "The network chain ID")

	fs.String(FlagIP, defaultIP, "The node's public IP")
	fs.String(FlagP2PPort, defaultP2PPort, "The node's P2P port")

	fs = addValidatorDescriptionFlags(fs)

	return fs
}

func NewFlagSetCollectGenVals(defaultHome string) *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(flags.FlagHome, defaultHome, "The application home directory")
	fs.String(FlagGenValDir, "", "Override default \"genval\" directory")

	return fs
}
