package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type rootConfig struct {
	APIKey   string
	BaseURL  string
	Provider string
}

var (
	cfg         rootConfig
	versionFlag bool
	rootCmd     = &cobra.Command{
		Use:     "sag",
		Short:   "🗣️ Multi-provider TTS, mac-style ease",
		Long:    "Command-line TTS with macOS playback. Supports ElevenLabs and Inworld providers.\nCall it like macOS 'say': if you skip the subcommand, text args are passed to 'speak' (e.g. `sag \"Hello\"`).\n\nTip: run `sag prompting` for model-specific prompting tips.\n\nProviders:\n  elevenlabs (default): eleven_v3, eleven_multilingual_v2, eleven_flash_v2_5\n  inworld: inworld-tts-1, inworld-tts-1-max",
		Example: "  sag \"Hi Peter\"\n  sag --provider inworld \"Hello from Inworld\"\n  echo 'piped input' | sag\n  sag speak -v Roger --rate 200 \"Faster speech\"",
		Version: "0.2.1",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if versionFlag {
				fmt.Println(cmd.Root().Name(), cmd.Root().Version)
				os.Exit(0)
			}
			return nil
		},
	}
)

// Execute is the entry point from main.
func Execute() {
	maybeDefaultToSpeak()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfg.Provider, "provider", "elevenlabs", "TTS provider (elevenlabs, inworld)")
	rootCmd.PersistentFlags().StringVar(&cfg.APIKey, "api-key", "", "API key (or provider-specific env var)")
	rootCmd.PersistentFlags().StringVar(&cfg.BaseURL, "base-url", "", "Override API base URL")
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "V", false, "Print version and exit")
}

// maybeDefaultToSpeak injects the "speak" subcommand when the user calls `sag` like macOS `say`.
func maybeDefaultToSpeak() {
	if len(os.Args) <= 1 {
		return
	}

	// npm/pnpm pass-through typically prefixes args with "--"; drop it so flags still parse.
	if os.Args[1] == "--" {
		os.Args = append([]string{os.Args[0]}, os.Args[2:]...)
		if len(os.Args) <= 1 {
			return
		}
	}

	first := os.Args[1]
	if isKnownSubcommand(first) || first == "-h" || first == "--help" {
		return
	}
	os.Args = append([]string{os.Args[0], "speak"}, os.Args[1:]...)
}

func isKnownSubcommand(name string) bool {
	name = strings.ToLower(name)
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return true
		}
		for _, a := range cmd.Aliases {
			if a == name {
				return true
			}
		}
	}
	return false
}
