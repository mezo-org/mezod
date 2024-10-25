package toml

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cosmossdk.io/tools/confix"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
	"github.com/creachadair/tomledit/transform"
	"github.com/spf13/cobra"
)

// Pattern for detecting a valid TOML array
//
// Good luck with understanding this ;)
//
// https://regex101.com/r/45Wetr/1
// This regex detects TOML array syntax.
//
// It supports depth 1 nested arrays and simple arrays:
// [["apple","banana"],["orange"]]
// ["apple","banana"]
// ["apple","banana",["orange","cherry"]]
//
// This is not supported for example (depth 2 nested array):
// ["apple","banana",["orange","cherry",["apple","banana"]]]
const tomlArrayPattern = `^\[\s*(?:(?:\[(?:[^\[\]]*)\])|(?:[^\[\],]+))\s*(?:,\s*(?:(?:\[(?:[^\[\]]*)\])|(?:[^\[\],]+))\s*)*\]$`

const SetTOMLCmdLong = `Set a specific part of toml configuration. You should specify path of the config file to be edited, ` +
	`section for the attribute, attribute name and desired value`

const SetTOMLCmdExample = "toml set <path_to_config_file> -v section.test=<desired_value> -v section.test2='[\"example\",\"array\"]'"

func NewSetCmd() *cobra.Command {
	var tomlValues []string
	var FlagStdOut bool

	cmd := &cobra.Command{
		Use:     "set",
		Short:   "Edit TOML configuration",
		Long:    SetTOMLCmdLong,
		Example: SetTOMLCmdExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			// read config file path from args
			configFile := args[0]
			isFile, err := isFile(configFile)
			if args[0] == "" || !isFile || err != nil {
				return errors.New("no valid path to file specified")
			}

			// map []string arguments from command line to custom TOMLKeyValue struct
			parsedTomlFromUser := parseTOMLKeyValues(tomlValues)

			// generate a transform plan for each value from config we want to change
			transformPlans := generateTransformPlans(parsedTomlFromUser, configFile)

			// if output path is set to empty string, the expected change will only be printed to stdout
			outputPath := configFile
			if FlagStdOut {
				outputPath = ""
			}

			// for our use case set skipValidate arg to true, otherwise some config files could not be edited (cometbft config)
			resultOfTransform := transformValues(context.Background(), transformPlans, configFile, outputPath, true)

			return resultOfTransform
		},
	}

	cmd.Flags().StringArrayVarP(&tomlValues, "value", "v", []string{}, "Define values to display/edit their values.")
	cmd.Flags().BoolVar(&FlagStdOut, "stdout", false, "print the updated config to stdout")

	return cmd
}

// wrapper function for confix.Upgrade that edits config file in-place
func transformValues(ctx context.Context, plans []transform.Plan, configPath string, outputPath string, skipValidate bool) error {
	var result error
	for _, plan := range plans {
		fmt.Println(plan[0].Desc)
		result = confix.Upgrade(ctx, plan, configPath, outputPath, skipValidate)
		if result != nil {
			fmt.Println("result contains errors: ", result)
			return result
		}
		fmt.Println("plan successfully applied")
	}

	return result
}

// function for checking if string has TOML array-like format (using regex)
func isValidTOMLArray(input string) bool {
	// Compile the regex pattern
	re := regexp.MustCompile(tomlArrayPattern)
	// Use MatchString to check if the input matches the pattern
	return re.MatchString(input)
}

// generate array of plans ([]transform.Plan) that define TOML variable change logic
func generateTransformPlans(tomlKeyValues []keyValueTOML, filename string) []transform.Plan {
	transformPlans := []transform.Plan{}

	for _, keyvalue := range tomlKeyValues {
		key, value := keyvalue.Key, keyvalue.Value
		plan := transform.Plan{
			{
				Desc: fmt.Sprintf("update %s = %s in %s", strings.Join(key, "."), value, filename),
				T: transform.Func(func(_ context.Context, doc *tomledit.Document) error {
					// depending on the value, we try to convert it to various formats, default is string
					var convertedValue string
					if isValidTOMLArray(value) {
						convertedValue = strings.ReplaceAll(value, "\\", "")
					} else if intVal, err := strconv.Atoi(value); err == nil {
						convertedValue = fmt.Sprintf("%d", intVal)
					} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
						convertedValue = strconv.FormatFloat(floatVal, 'f', -1, 64)
					} else if boolVal, err := strconv.ParseBool(value); err == nil {
						convertedValue = fmt.Sprintf("%t", boolVal)
					} else {
						convertedValue = fmt.Sprintf("%q", value)
					}

					// finding the section containing entries []*tomledit.Entry
					found := doc.Find(key...)
					switch {
					case len(found) == 0:
						return fmt.Errorf("key %q not found", key)
					case len(found) > 1:
						return fmt.Errorf("found %d definitions of key %q", len(found), key)
					case !found[0].IsMapping():
						return fmt.Errorf("%q is not a key-value mapping", key)
					}

					// convert value to parser.Value type - transform.InsertMapping requires it
					parserValue, err := parser.ParseValue(convertedValue)
					if err != nil {
						return err
					}

					// validate transform
					if ok := transform.InsertMapping(found[0].Section, &parser.KeyValue{
						Block: found[0].Block,
						Name:  found[0].Name,
						Value: parserValue,
					}, true); !ok {
						return errors.New("failed to set value")
					}
					return nil
				}),
			},
		}

		transformPlans = append(transformPlans, plan)
	}

	return transformPlans
}

func parseTOMLKeyValues(tomlValues []string) []keyValueTOML {
	parsedTOML := []keyValueTOML{}

	for _, val := range tomlValues {
		var newTOMLEntry keyValueTOML

		splitKVString := strings.SplitN(val, "=", 2)
		if len(splitKVString) != 2 {
			fmt.Printf("Skipping invalid key-value pair: %s\n", val)
			continue
		}

		inputKey := strings.TrimSpace(splitKVString[0])
		inputValue := strings.TrimSpace(splitKVString[1])
		inputKeySections := strings.Split(inputKey, ".")

		newTOMLEntry.Key = inputKeySections
		newTOMLEntry.Value = inputValue

		parsedTOML = append(parsedTOML, newTOMLEntry)
	}

	return parsedTOML
}
