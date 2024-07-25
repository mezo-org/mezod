package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/evmos/evmos/v12/x/poa/types"
)

// NewQueryCmd returns the cli query commands for this module
func NewQueryCmd() *cobra.Command {
	// Group poa queries under a subcommand
	poaQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	poaQueryCmd.AddCommand(
		NewCmdQueryValidator(),
		NewCmdQueryValidators(),
		NewCmdQueryParams(),
		NewCmdQueryApplications(),
	)

	return poaQueryCmd
}

// NewCmdQueryValidator queries information about a validator
func NewCmdQueryValidator() *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryValidatorRequest{
				Operator: args[0],
			}

			response, err := queryClient.Validator(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
}

// NewCmdQueryValidators queries all validators
func NewCmdQueryValidators() *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "Query all validators",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryValidatorsRequest{}

			response, err := queryClient.Validators(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
}

// NewCmdQueryParams queries the params
func NewCmdQueryParams() *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Short: "Query the params",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryParamsRequest{}

			response, err := queryClient.Params(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
}

// NewCmdQueryApplications queries the applications to become a validator
func NewCmdQueryApplications() *cobra.Command {
	return &cobra.Command{
		Use:   "applications",
		Short: "Query the applications to become validator",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			request := &types.QueryApplicationsRequest{}

			response, err := queryClient.Applications(context.Background(), request)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(response)
		},
	}
}
