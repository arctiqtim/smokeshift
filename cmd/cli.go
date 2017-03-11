package main

import (
	"io"
	"github.com/opencredo/smokeshift/pkg/config"
	"github.com/opencredo/smokeshift/pkg/smokeshift"
	"github.com/spf13/cobra"

)

// NewKismaticCommand creates the kismatic command
func NewSmokeshiftCommand(version string, in io.Reader, out io.Writer) *cobra.Command {
	var skipCleanup bool
	cmd := &cobra.Command{
		Use:   "smokeshift",
		SilenceUsage:  true,
		SilenceErrors: true,
		Short: "smokeshift tests your Openshift cluster using the oc CLI",
		Long: `smokeshift is intended to perform smoke tests against an Openshift cluster. It expects the oc cli
to be available on the path and a user, with cluster-admin access, to already have authenticated. The actual smoke test
creates a Project (smoketest), deploys an Nginx Pod on to each Node, provisions a busybox and uses that to ensure access
using DNS and IP based connections to the Nginx Pods. Unless the 'skip-cleanup' flag is set all Pods, Services and the
smokeshift Project are deleted on completion`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return doCheckOpenshift(skipCleanup)
		},

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
