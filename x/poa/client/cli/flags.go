package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagAddressValidator = "validator"

	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"
)

// common flagsets to add to various functions
var (
	fsValidator = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
}

func FlagSetDescriptionCreate() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	fs.String(FlagMoniker, "", "The validator's name")
	fs.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fs.String(FlagWebsite, "", "The validator's (optional) website")
	fs.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fs.String(FlagDetails, "", "The validator's (optional) details")

	return fs
}
