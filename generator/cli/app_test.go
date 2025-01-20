package cli_test

import (
	"testing"

	"github.com/EliCDavis/polyform/generator/cli"
	"github.com/stretchr/testify/assert"
)

func TestRunCallsCommand(t *testing.T) {
	// ARRANGE ================================================================
	var values []string
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:    "command",
				Aliases: []string{"command", "c"},
				Run: func(appState *cli.RunState) error {
					values = appState.Args
					return nil
				},
			},
		},
	}

	// ACT ====================================================================
	err := app.Run([]string{"testProgram", "command", "1", "2"})

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Len(t, values, 2)
	assert.Equal(t, "1", values[0])
	assert.Equal(t, "2", values[1])
}

func TestRunCallsHelpOnUnknownCommand(t *testing.T) {
	// ARRANGE ================================================================
	called := false
	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:    "Help",
				Aliases: []string{"help"},
				Run: func(appState *cli.RunState) error {
					called = true
					return nil
				},
			},
		},
	}

	// ACT ====================================================================
	err := app.Run([]string{"testProgram"})

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.True(t, called)
}
