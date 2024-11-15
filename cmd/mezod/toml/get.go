package toml

import (
	"errors"
	"fmt"
	"strings"

	"cosmossdk.io/tools/confix"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/spf13/cobra"
)

const GetTOMLCmdLong = `Get a specific part of toml configuration. You should specify path of the config file to be edited, ` +
	`section for the attribute and an attribute name.`

const GetTOMLCmdExample = "toml get <path_to_config_file> -v section.test=<desired_value>"

func NewGetCmd() *cobra.Command {
	var tomlInputKeys []string
	cmd := &cobra.Command{
		Use:     "get",
		Short:   "Display TOML configuration",
		Long:    GetTOMLCmdLong,
		Example: GetTOMLCmdExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// read config file path from args
			configFile := args[0]
			isFile, err := isFile(configFile)
			if args[0] == "" || !isFile || err != nil {
				return errors.New("no valid path to file specified")
			}
			// load toml document using confix
			doc, err := confix.LoadConfig(configFile)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// parse flag for verbose toml key printing
			printKeyFlag, err := cmd.Flags().GetBool("print-key")
			if err != nil {
				return err
			}

			// parse flag for printing all toml key-values
			isAll, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}

			if isAll {
				return printFullConfig(doc, printKeyFlag)
			}

			// parse toml keys to keyValueTOML data structure
			// we use TOMLKeyValue, but all values are empty strings
			tomlKeys := parseTOMLKeys(tomlInputKeys)
			for _, keys := range tomlKeys {
				results := doc.Find(keys.Key...)
				if len(results) == 0 {
					return fmt.Errorf("key %v not found", keys.Key)
				} else if len(results) > 1 {
					return fmt.Errorf("key %v is ambiguous", keys.Key)
				}

				if printKeyFlag {
					fmt.Printf("%s\n", results[0].KeyValue)
				} else {
					fmt.Printf("%s\n", results[0].KeyValue.Value)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP("all", "a", false, "View all values in the configuration")
	cmd.Flags().BoolP("print-key", "p", false, "Print keys of given values instead of raw value")
	cmd.Flags().StringArrayVarP(&tomlInputKeys, "value", "v", []string{}, "View given values")

	return cmd
}

func parseTOMLKeys(tomlValues []string) []keyValueTOML {
	parsedTOML := []keyValueTOML{}

	for _, val := range tomlValues {
		var newTOMLEntry keyValueTOML

		splitKeyStructure := strings.Split(val, ".")
		if len(splitKeyStructure) < 1 {
			fmt.Println("no value for reading specified")
			return nil
		}

		// we don't need the value here
		newTOMLEntry.Key = splitKeyStructure
		newTOMLEntry.Value = ""

		parsedTOML = append(parsedTOML, newTOMLEntry)
	}

	return parsedTOML
}

// i know passing boolean into a function isn't the best practice but now i think it's justified
func printFullConfig(doc *tomledit.Document, printKeyFlag bool) error {
	// Print global key-value pairs (if any) at the root level
	if doc.Global != nil {
		for _, item := range doc.Global.Items {
			if keyValueItem, ok := item.(*parser.KeyValue); ok {
				if printKeyFlag {
					fmt.Println(keyValueItem.String())
				} else {
					fmt.Println(keyValueItem.Value)
				}
			}
		}
	}

	// Print sectioned key-value pairs
	for _, s := range doc.Sections {
		fmt.Println(s.String())
		for _, item := range s.Items {
			if keyValueItem, ok := item.(*parser.KeyValue); ok {
				if printKeyFlag {
					fmt.Println(keyValueItem.String())
				} else {
					fmt.Println(keyValueItem.Value)
				}
			}
		}
	}
	return nil
}
