package commands

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
	"github.com/vitalvas/oneauth/internal/yubikey"
)

// skipIfNoPCSC skips the test if PC/SC is not available
func skipIfNoPCSC(t *testing.T) {
	t.Helper()
	_, err := yubikey.Cards()
	if err != nil {
		errStr := err.Error()
		// Check for various PC/SC related errors
		if strings.Contains(errStr, "Smart card resource manager is not running") ||
			strings.Contains(errStr, "connecting to pscs") ||
			strings.Contains(errStr, "the Smart card resource manager is not running") ||
			strings.Contains(errStr, "failed to list cards") {
			t.Skipf("Skipping test: PC/SC service is not available: %v", err)
		}
	}
}

// skipIfYubiKeyConnected skips the test if a YubiKey is connected
// Used for tests that assume no YubiKey is present
func skipIfYubiKeyConnected(t *testing.T) {
	t.Helper()
	cards, err := yubikey.Cards()
	if err != nil {
		// If we can't list cards, that's fine - no YubiKey detected
		return
	}
	if len(cards) > 0 {
		t.Skipf("Skipping test: YubiKey is connected (found %d card(s))", len(cards))
	}
}

func TestSelectYubiKey(t *testing.T) {
	t.Run("NoYubiKeys", func(t *testing.T) {
		skipIfYubiKeyConnected(t)

		// Create a mock context
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 0, "")

		ctx := cli.NewContext(app, set, nil)

		// This will fail because no YubiKeys are connected
		err := selectYubiKey(ctx)
		assert.Error(t, err)
	})

	t.Run("SerialSetButNotFound", func(t *testing.T) {
		skipIfNoPCSC(t)

		// Create a mock context with a serial that won't be found
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 999999, "")

		ctx := cli.NewContext(app, set, nil)

		// This will fail because the specified serial won't be found
		err := selectYubiKey(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YubiKey with serial 999999 not found")
	})

	t.Run("ErrorMessages", func(t *testing.T) {
		tests := []struct {
			name        string
			serial      uint64
			expectedErr string
		}{
			{
				name:        "SerialNotFound",
				serial:      123456,
				expectedErr: "YubiKey with serial 123456 not found",
			},
			{
				name:        "AnotherSerialNotFound",
				serial:      999999,
				expectedErr: "YubiKey with serial 999999 not found",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				skipIfNoPCSC(t)
				app := &cli.App{
					Flags: []cli.Flag{
						&cli.Uint64Flag{Name: "serial"},
					},
				}

				set := flag.NewFlagSet("test", 0)
				set.Uint64("serial", tt.serial, "")

				ctx := cli.NewContext(app, set, nil)

				err := selectYubiKey(ctx)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			})
		}
	})
}

func TestSelectYubiKeyContextHandling(t *testing.T) {
	t.Run("ContextSetup", func(t *testing.T) {
		skipIfYubiKeyConnected(t)

		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 0, "")

		ctx := cli.NewContext(app, set, nil)

		// Verify initial state
		assert.Equal(t, uint64(0), ctx.Uint64("serial"))

		// The function should fail because no YubiKeys are available
		err := selectYubiKey(ctx)
		assert.Error(t, err)
	})

	t.Run("SerialFlagHandling", func(t *testing.T) {
		skipIfYubiKeyConnected(t)

		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		testCases := []uint64{0, 123456, 999999}

		for _, serial := range testCases {
			t.Run(fmt.Sprintf("Serial_%d", serial), func(t *testing.T) {
				set := flag.NewFlagSet("test", 0)
				set.Uint64("serial", serial, "")

				ctx := cli.NewContext(app, set, nil)

				assert.Equal(t, serial, ctx.Uint64("serial"))

				// Function should error due to no YubiKeys
				err := selectYubiKey(ctx)
				assert.Error(t, err)
			})
		}
	})
}

func TestSelectYubiKeyLogic(t *testing.T) {
	t.Run("ZeroSerialHandling", func(t *testing.T) {
		skipIfYubiKeyConnected(t)

		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 0, "")

		ctx := cli.NewContext(app, set, nil)

		// When serial is 0, it should try to find cards
		err := selectYubiKey(ctx)
		assert.Error(t, err)
		// Should get error about card count, not serial not found
		assert.NotContains(t, err.Error(), "YubiKey with serial")
	})

	t.Run("NonZeroSerialHandling", func(t *testing.T) {
		skipIfNoPCSC(t)
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 123456, "")

		ctx := cli.NewContext(app, set, nil)

		// When serial is non-zero, it should look for specific card
		err := selectYubiKey(ctx)
		assert.Error(t, err)
		// Should get error about serial not found
		assert.Contains(t, err.Error(), "YubiKey with serial 123456 not found")
	})
}

func TestSelectYubiKeyEdgeCases(t *testing.T) {
	t.Run("MaxUint64Serial", func(t *testing.T) {
		skipIfNoPCSC(t)
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", ^uint64(0), "") // Max uint64

		ctx := cli.NewContext(app, set, nil)

		err := selectYubiKey(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "YubiKey with serial")
	})

	t.Run("LargeSerialNumbers", func(t *testing.T) {
		largeSerials := []uint64{
			1000000,
			4294967295, // Max uint32
			4294967296, // Max uint32 + 1
		}

		for _, serial := range largeSerials {
			t.Run(fmt.Sprintf("Serial_%d", serial), func(t *testing.T) {
				skipIfNoPCSC(t)
				app := &cli.App{
					Flags: []cli.Flag{
						&cli.Uint64Flag{Name: "serial"},
					},
				}

				set := flag.NewFlagSet("test", 0)
				set.Uint64("serial", serial, "")

				ctx := cli.NewContext(app, set, nil)

				err := selectYubiKey(ctx)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), fmt.Sprintf("YubiKey with serial %d not found", serial))
			})
		}
	})
}

// Test helper function behavior
func TestSelectYubiKeyFunctionBehavior(t *testing.T) {
	t.Run("FunctionSignature", func(t *testing.T) {
		skipIfYubiKeyConnected(t)

		// Test that function accepts cli.Context and returns error
		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 0, "")

		ctx := cli.NewContext(app, set, nil)

		// Function should return an error (not nil)
		err := selectYubiKey(ctx)
		assert.Error(t, err)
		assert.NotNil(t, err)
	})

	t.Run("ErrorsAreProperlyFormatted", func(t *testing.T) {
		skipIfNoPCSC(t)

		app := &cli.App{
			Flags: []cli.Flag{
				&cli.Uint64Flag{Name: "serial"},
			},
		}

		set := flag.NewFlagSet("test", 0)
		set.Uint64("serial", 123456, "")

		ctx := cli.NewContext(app, set, nil)

		err := selectYubiKey(ctx)
		assert.Error(t, err)
		assert.NotEmpty(t, err.Error())
		assert.Implements(t, (*error)(nil), err)
	})
}
