package cli

import (
	"bufio"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	poaTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poaTxCmd.AddCommand(flags.PostCommands(
		GetCmdSubmitApplication(cdc),
		GetCmdProposeKick(cdc),
		GetCmdVoteApplication(cdc),
		GetCmdVoteKickProposal(cdc),
		GetCmdLeaveValidatorSet(cdc),
	)...)

	return poaTxCmd
}

// GetCmdSubmitApplication sends a new application to become a validator
func GetCmdSubmitApplication(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [validator-consensus-pubkey]",
		Short: "Apply to become a new validator in the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Operator address is the sender
			accAddress := cliCtx.GetFromAddress()
			if accAddress.Empty() {
				return fmt.Errorf("Account address empty")
			}

			opAddress := sdk.ValAddress(accAddress)

			// Consensus public key for the validator
			pk, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, args[0])
			if err != nil {
				return fmt.Errorf("Cannot convert pubkey: %v", err)
			}

			// Description of the candidate
			moniker, _ := cmd.Flags().GetString(FlagMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)
			description := types.NewDescription(moniker, identity, website, security, details)

			candidateValidator := types.NewValidator(opAddress, pk, description)

			msg := types.NewMsgSubmitApplication(candidateValidator)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	cmd.Flags().AddFlagSet(FlagSetDescriptionCreate())

	return cmd
}

// GetCmdProposeKick sends a new kick proposal to remove a validator
func GetCmdProposeKick(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "propose-kick [validator-addr]",
		Short: "Propose to kick a validator from the validator",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Proposer address is the sender
			accAddress := cliCtx.GetFromAddress()
			if accAddress.Empty() {
				return fmt.Errorf("Account address empty")
			}
			proposeAddress := sdk.ValAddress(accAddress)

			// Get candidate address
			candidateAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgProposeKick(candidateAddr, proposeAddress)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdVoteApplication approves or rejects an application to become validator
func GetCmdVoteApplication(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote-application [candidate-addr] approve|reject",
		Short: "Approve or reject the application to become a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Voter address is the sender
			accAddress := cliCtx.GetFromAddress()
			if accAddress.Empty() {
				return fmt.Errorf("Account address empty")
			}
			voterAddress := sdk.ValAddress(accAddress)

			// Get candidate address
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Check if approved or rejected
			var approved bool
			if args[1] == "approve" {
				approved = true
			} else if args[1] == "reject" {
				approved = false
			} else {
				return fmt.Errorf("Vote neither approved nor rejected")
			}

			msg := types.NewMsgVote(types.VoteTypeApplication, voterAddress, valAddr, approved)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdVoteKickProposal approves or rejects a kick proposal
func GetCmdVoteKickProposal(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "vote-kick-proposal [candidate-addr] approve|reject",
		Short: "Approve or reject a kick proposal to remove a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Voter address is the sender
			accAddress := cliCtx.GetFromAddress()
			if accAddress.Empty() {
				return fmt.Errorf("Account address empty")
			}
			voterAddress := sdk.ValAddress(accAddress)

			// Get candidate address
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Check if approved or rejected
			var approved bool
			if args[1] == "approve" {
				approved = true
			} else if args[1] == "reject" {
				approved = false
			} else {
				return fmt.Errorf("Vote neither approved nor rejected")
			}

			msg := types.NewMsgVote(types.VoteTypeKickProposal, voterAddress, valAddr, approved)
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdLeaveValidatorSet remove oneself from the validator set
func GetCmdLeaveValidatorSet(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "leave-validator-set",
		Short: "Instantly leave the validator set",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))

			// Validator address is the sender
			accAddress := cliCtx.GetFromAddress()
			if accAddress.Empty() {
				return fmt.Errorf("Account address empty")
			}
			validatorAddress := sdk.ValAddress(accAddress)

			msg := types.NewMsgLeaveValidatorSet(validatorAddress)
			err := msg.ValidateBasic()
			if err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}
