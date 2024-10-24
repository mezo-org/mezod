// Copyright 2022 Evmos Foundation
// This file is part of the Evmos Network packages.
//
// Evmos is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The Evmos packages are distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the Evmos packages. If not, see https://github.com/evmos/evmos/blob/main/LICENSE

package toml

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const SetTOMLCmdLong = `Set a specific part of toml configuration. You should specify path of the config file to be edited, ` +
	`section for the attribute, attribute name and desired value`

const SetTOMLCmdExample = "toml set section.dupa 1234 --file=/var/mezod/(config.toml/app.toml/client.toml)"

const GetTOMLCmdLong = `Get a specific part of toml configuration. You should specify path of the config file to be edited, ` +
	`section for the attribute and an attribute name.`

const GetTOMLCmdExample = "toml get section.dupa --file=/var/mezod/(config.toml/app.toml/client.toml)"

// TOML_SECTION.TOML_ATTRIBUTE VALUE --path=<CONFIG_FILE_PATH>
func NewSetTOML() *cobra.Command {
	var tomlValues []string

	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Edit or display TOML configuration",
		Long:    SetTOMLCmdLong,
		Example: SetTOMLCmdExample,
		// Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// fmt.Println("args:", args)

			// file, _ := cmd.Flags().GetString("file")
			// values, _ := cmd.Flags().GetString("values")

			// fmt.Println(values)
			// fmt.Println(file)

			// return nil

			keyValuePairs := make(map[string]string)
			for _, val := range tomlValues {
				// Split using the first occurrence of '=' only
				parts := strings.SplitN(val, "=", 2)
				if len(parts) == 2 {
					key := parts[0]
					value := parts[1]
					keyValuePairs[key] = value
				} else {
					fmt.Printf("Invalid format for key-value pair: %s\n", val)
				}
			}

			// Output the key-value pairs
			fmt.Println("Parsed key-value pairs:")
			for key, value := range keyValuePairs {
				fmt.Printf("  %s = %s\n", key, value)
			}

			return nil
		},
	}

	cmd.Flags().String("file", "~/.mezod/config/app.toml", "Path to file you want to edit/get values from.")
	cmd.Flags().StringSliceVarP(&tomlValues, "value", "v", []string{}, "Define values to display/edit their values.")

	return cmd
}

// func NewGetTOML() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "toml ",
// 		Short: "Edit or display TOML configuration",
// 	}
// }
