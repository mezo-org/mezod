package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/tx"
	//nolint:staticcheck
	"github.com/cosmos/cosmos-sdk/types/bech32/legacybech32"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// NewTxCmd returns the transaction commands for this module
func NewTxCmd() *cobra.Command {
	poaTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poaTxCmd.AddCommand(
		NewCmdSubmitApplication(),
		NewCmdProposeKick(),
		NewCmdVoteApplication(),
		NewCmdVoteKickProposal(),
		NewCmdLeaveValidatorSet(),
	)

	return poaTxCmd
}

// NewCmdSubmitApplication sends a new application to become a validator
func NewCmdSubmitApplication() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [validator-consensus-pubkey]",
		Short: "Apply to become a new validator in the network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Operator address is the sender
			accAddr := clientCtx.GetFromAddress()
			if accAddr.Empty() {
				return fmt.Errorf("account address empty")
			}
			operatorAddr := sdk.ValAddress(accAddr)

			// Consensus public key for the validator
			//nolint:staticcheck
			publicKey, err := legacybech32.UnmarshalPubKey(
				legacybech32.ConsPK,
				args[0],
			)
			if err != nil {
				return fmt.Errorf("cannot convert pubkey: %v", err)
			}

			// Description of the candidate
			moniker, _ := cmd.Flags().GetString(FlagMoniker)
			identity, _ := cmd.Flags().GetString(FlagIdentity)
			website, _ := cmd.Flags().GetString(FlagWebsite)
			security, _ := cmd.Flags().GetString(FlagSecurityContact)
			details, _ := cmd.Flags().GetString(FlagDetails)
			description := types.NewDescription(moniker, identity, website, security, details)

			candidate := types.NewValidator(operatorAddr, publicKey, description)

			msg := &types.MsgSubmitApplication{
				Candidate: candidate,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().AddFlagSet(NewFlagSetSubmitApplication())

	return cmd
}

// NewCmdProposeKick sends a new kick proposal to remove a validator
func NewCmdProposeKick() *cobra.Command {
	return &cobra.Command{
		Use:   "propose-kick [validator-addr]",
		Short: "Propose to kick a validator from the validators pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Proposer address is the sender
			accAddr := clientCtx.GetFromAddress()
			if accAddr.Empty() {
				return fmt.Errorf("account address empty")
			}
			proposerAddr := sdk.ValAddress(accAddr)

			// Get candidate address
			candidateAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := &types.MsgProposeKick{
				CandidateAddr: candidateAddr,
				ProposerAddr:  proposerAddr,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewCmdVoteApplication approves or rejects an application to become validator
func NewCmdVoteApplication() *cobra.Command {
	return &cobra.Command{
		Use:   "vote-application [candidate-addr] approve|reject",
		Short: "Approve or reject the application to become a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Voter address is the sender
			accAddr := clientCtx.GetFromAddress()
			if accAddr.Empty() {
				return fmt.Errorf("account address empty")
			}
			voterAddr := sdk.ValAddress(accAddr)

			// Get candidate address
			candidateAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Check if approved or rejected
			var approved bool
			switch args[1] {
			case "approve":
				approved = true
			case "reject":
				approved = false
			default:
				return fmt.Errorf("vote neither approved nor rejected")
			}

			msg := &types.MsgVote{
				VoteType:      types.VoteTypeApplication,
				VoterAddr:     voterAddr,
				CandidateAddr: candidateAddr,
				Approve:       approved,
			}
			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewCmdVoteKickProposal approves or rejects a kick proposal
func NewCmdVoteKickProposal() *cobra.Command {
	return &cobra.Command{
		Use:   "vote-kick-proposal [candidate-addr] approve|reject",
		Short: "Approve or reject a kick proposal to remove a validator",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Voter address is the sender
			accAddr := clientCtx.GetFromAddress()
			if accAddr.Empty() {
				return fmt.Errorf("account address empty")
			}
			voterAddr := sdk.ValAddress(accAddr)

			// Get candidate address
			candidateAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// Check if approved or rejected
			var approved bool
			switch args[1] {
			case "approve":
				approved = true
			case "reject":
				approved = false
			default:
				return fmt.Errorf("vote neither approved nor rejected")
			}

			msg := &types.MsgVote{
				VoteType:      types.VoteTypeKickProposal,
				VoterAddr:     voterAddr,
				CandidateAddr: candidateAddr,
				Approve:       approved,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

// NewCmdLeaveValidatorSet remove oneself from the validator set
func NewCmdLeaveValidatorSet() *cobra.Command {
	return &cobra.Command{
		Use:   "leave-validator-set",
		Short: "Instantly leave the validator set",
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Validator address is the sender
			accAddr := clientCtx.GetFromAddress()
			if accAddr.Empty() {
				return fmt.Errorf("account address empty")
			}
			validatorAddr := sdk.ValAddress(accAddr)

			msg := &types.MsgLeaveValidatorSet{
				ValidatorAddr: validatorAddr,
			}

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
