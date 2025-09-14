package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/krjakbrjak/name-resolver/pkg/resolver"
	"github.com/spf13/cobra"
)

const (
	DEBUG = "debug"
	INFO  = "info"
	WARN  = "warn"
	ERROR = "error"
)

func getLogLevel() slog.Level {
	level, ok := os.LookupEnv("LOG_LEVEL")
	if !ok {
		level = INFO
	}
	switch strings.ToLower(level) {
	case DEBUG:
		return slog.LevelDebug
	case INFO:
		return slog.LevelInfo
	case WARN:
		return slog.LevelWarn
	case ERROR:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// dnsCmd represents the dns command
var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Starts a DNS server that resolves container names and labels to IP addresses.",
	Long: `The command inspects running Docker containers, applies the specified filters and mappings,
and serves DNS responses accordingly. If a query cannot be resolved locally, it is forwarded
to the specified fallback DNS servers. Logging is provided to stdout with configurable log level.
`,
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

		mappingEntries, mapErr := cmd.Flags().GetStringArray("map")
		if mapErr != nil {
			return mapErr
		}

		mapping := make(map[string]string)
		for _, entry := range mappingEntries {
			parts := strings.SplitN(entry, ":", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid map entry: %s", entry)
			}
			mapping[parts[0]] = parts[1]
		}

		inspector, inspectorErr := resolver.NewDockerInspector()
		if inspectorErr != nil {
			return inspectorErr
		}

		logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: getLogLevel()}))
		return resolver.Serve(port, resolver.Filter{
			Name:    name,
			Labels:  labels,
			Mapping: mapping,
		}, fallbackDns, inspector, logger)
	},
}

func init() {
	rootCmd.AddCommand(dnsCmd)

	dnsCmd.Flags().Uint16P("port", "p", 53, "DNS port")
	dnsCmd.Flags().StringArrayP("label", "l", []string{}, "Containers labels")
	dnsCmd.Flags().StringP("name", "n", "", "Containers name filter")
	dnsCmd.Flags().StringArrayP("map", "m", []string{}, "Hostname to container name mapping in format hostname:container_name")
	dnsCmd.Flags().StringArrayP("fallback-dns", "d", []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
		"8.8.4.4:53",
		"1.0.0.1:53",
	}, "List of fallback DNS addresses")
}
