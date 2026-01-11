package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/steipete/sag/internal/tts"

	// Import providers to register them
	_ "github.com/steipete/sag/internal/elevenlabs"
	_ "github.com/steipete/sag/internal/inworld"

	"github.com/spf13/cobra"
)

type voicesOptions struct {
	search string
	limit  int
}

func init() {
	opts := voicesOptions{
		limit: 100,
	}

	cmd := &cobra.Command{
		Use:   "voices",
		Short: "List available voices for the current provider",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			return ensureAPIKey()
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			provider, err := tts.Get(cfg.Provider, cfg.APIKey, cfg.BaseURL)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()

			voices, err := provider.ListVoices(ctx, opts.search)
			if err != nil {
				return err
			}

			if opts.limit > 0 && len(voices) > opts.limit {
				voices = voices[:opts.limit]
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			if _, err := fmt.Fprintf(w, "VOICE ID\tNAME\tCATEGORY\n"); err != nil {
				return err
			}
			for _, v := range voices {
				if _, err := fmt.Fprintf(w, "%s\t%s\t%s\n", v.ID, v.Name, v.Category); err != nil {
					return err
				}
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&opts.search, "search", "", "Filter voices by name (server-side when supported)")
	cmd.Flags().IntVar(&opts.limit, "limit", opts.limit, "Maximum rows to display (0 = all)")
	rootCmd.AddCommand(cmd)
}
