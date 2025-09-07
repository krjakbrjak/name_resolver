package cmd

import (
	"github.com/krjakbrjak/name-resolver/pkg/resolver"
	"github.com/spf13/cobra"
)

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, portErr := cmd.Flags().GetUint16("port")
		if portErr != nil {
			return portErr
		}

		name, nameErr := cmd.Flags().GetString("name")
		if nameErr != nil {
			return nameErr
		}

		labels, labelsErr := cmd.Flags().GetStringArray("label")
		if labelsErr != nil {
			return labelsErr
		}

		fallbackDns, fallbackDnsErr := cmd.Flags().GetStringArray("fallback-dns")
		if fallbackDnsErr != nil {
			return fallbackDnsErr
		}

		inspector, inspectorErr := resolver.NewDockerInspector()
		if inspectorErr != nil {
			return inspectorErr
		}

		return resolver.Serve(port, resolver.Filter{
			Name:   name,
			Labels: labels,
		}, fallbackDns, inspector)
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	dnsCmd.Flags().Uint16P("port", "p", 53, "DNS port")
	dnsCmd.Flags().StringArrayP("label", "l", []string{}, "Containers labels")
	dnsCmd.Flags().StringP("name", "n", "", "Containers name filter")
	dnsCmd.Flags().StringArrayP("fallback-dns", "d", []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
		"8.8.4.4:53",
		"1.0.0.1:53",
	}, "List of fallback DNS addresses")
}
