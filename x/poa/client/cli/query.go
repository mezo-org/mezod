package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/evmos/evmos/v12/x/poa/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	// Group poa queries under a subcommand
	poaQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poaQueryCmd.AddCommand(
		flags.GetCommands(
			GetCmdQueryValidator(queryRoute, cdc),
			GetCmdQueryValidators(queryRoute, cdc),
			GetCmdQueryParams(queryRoute, cdc),
			GetCmdQueryApplications(queryRoute, cdc),
			GetCmdQueryKickProposals(queryRoute, cdc),
		)...,
	)

	return poaQueryCmd
}

// GetCmdQueryValidator queries information about a validator
func GetCmdQueryValidator(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// Get address
			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			// We don't want pagination for cli queries
			params := types.NewQueryValidatorParams(addr)

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidator), bz)
			if err != nil {
				fmt.Printf("could not resolve %s %s \n", types.QueryValidator, addr)
				return nil
			}

			var out types.Validator
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryValidators queries all validators
func GetCmdQueryValidators(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "Query all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidators), nil)
			if err != nil {
				fmt.Printf("could not resolve %s \n", types.QueryValidators)
				return nil
			}

			var out []types.Validator
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryParams queries the params
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the params",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryParams), nil)
			if err != nil {
				fmt.Printf("could not resolve %s \n", types.QueryParams)
				return nil
			}

			var out types.Params
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryApplications queries the applications to become a validator
func GetCmdQueryApplications(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "applications",
		Short: "Query the applications to become validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryApplications), nil)
			if err != nil {
				fmt.Printf("could not resolve %s \n", types.QueryApplications)
				return nil
			}

			var out []types.Vote
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// GetCmdQueryKickProposals queries the kick proposals to remove a validator
func GetCmdQueryKickProposals(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "kick-proposals",
		Short: "Query the kick proposals to remove validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryKickProposals), nil)
			if err != nil {
				fmt.Printf("could not resolve %s \n", types.QueryKickProposals)
				return nil
			}

			var out []types.Vote
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
