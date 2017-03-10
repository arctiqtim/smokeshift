package main

import (
	"io"
	"github.com/cyberbliss/smokeshift/pkg/config"
	"github.com/cyberbliss/smokeshift/pkg/smokeshift"
	"github.com/spf13/cobra"

)

// NewKismaticCommand creates the kismatic command
func NewSmokeshiftCommand(version string, in io.Reader, out io.Writer) *cobra.Command {
	var skipCleanup bool
	cmd := &cobra.Command{
		Use:   "smokeshift",
		Short: "smokeshift tests your kubernetes cluster using kubectl",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doCheckOpenshift(skipCleanup)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	config.Namespace = "smokeshift"

	cmd.PersistentFlags().StringVar(&config.RegistryURL, "registry-url", "",
		"Override the default Docker Hub URL to use a local offline registry for required Docker images.")
	cmd.Flags().BoolVar(&skipCleanup, "skip-cleanup", false, "Don't clean up. Leave all deployed artifacts running on the cluster.")

	return cmd
}

func doCheckOpenshift(skipCleanup bool) error {
	return smokeshift.CheckOpenshift(skipCleanup)
}
