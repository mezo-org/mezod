package maintenance

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNextUpgradeName(t *testing.T) {
	noHandlers := func(string) bool { return false }
	allHandlers := func(string) bool { return true }

	// handlersFor reports a registered handler for the given names only.
	handlersFor := func(names ...string) func(string) bool {
		registered := make(map[string]bool, len(names))
		for _, name := range names {
			registered[name] = true
		}

		return func(name string) bool { return registered[name] }
	}

	// The binary carries a handler for every historical upgrade. A dev chain
	// that never applied them still cannot halt on those names.
	historical := make([]string, 0, 13)
	for major := 1; major <= 13; major++ {
		historical = append(historical, fmt.Sprintf("v%d.0.0", major))
	}

	tests := map[string]struct {
		lastCompleted string
		hasHandler    func(string) bool
		expectedName  string
		// expectedErr is the exact error text. The cap case leaves it empty
		// because the contract does not fix that text. It sets expectsErr.
		expectedErr string
		expectsErr  bool
	}{
		"minor version zero": {
			lastCompleted: "v12.0.0",
			hasHandler:    noHandlers,
			expectedName:  "v13.0.0",
		},
		"major version zero": {
			lastCompleted: "v0.4.0",
			hasHandler:    noHandlers,
			expectedName:  "v1.0.0",
		},
		"non-zero minor version": {
			lastCompleted: "v9.1.0",
			hasHandler:    noHandlers,
			expectedName:  "v10.0.0",
		},
		"empty name": {
			lastCompleted: "",
			hasHandler:    noHandlers,
			expectedName:  "v1.0.0",
		},
		"the first candidate has a handler": {
			lastCompleted: "",
			hasHandler:    handlersFor("v1.0.0"),
			expectedName:  "v2.0.0",
		},
		"every historical candidate has a handler": {
			lastCompleted: "v0.4.0",
			hasHandler:    handlersFor(historical...),
			expectedName:  "v14.0.0",
		},
		"every candidate has a handler": {
			lastCompleted: "v1.0.0",
			hasHandler:    allHandlers,
			expectsErr:    true,
		},
		"name is not a version": {
			lastCompleted: "garbage",
			hasHandler:    noHandlers,
			expectedErr:   `cannot derive the halt name from the last completed upgrade "garbage"`,
		},
		"version without the patch part": {
			lastCompleted: "v1.2",
			hasHandler:    noHandlers,
			expectedErr:   `cannot derive the halt name from the last completed upgrade "v1.2"`,
		},
		"version without the v prefix": {
			lastCompleted: "1.2.3",
			hasHandler:    noHandlers,
			expectedErr:   `cannot derive the halt name from the last completed upgrade "1.2.3"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			actualName, err := nextUpgradeName(test.lastCompleted, test.hasHandler)

			if test.expectsErr || test.expectedErr != "" {
				require.Error(t, err)
				if test.expectedErr != "" {
					require.EqualError(t, err, test.expectedErr)
				}
				require.Empty(t, actualName)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expectedName, actualName)
		})
	}
}
