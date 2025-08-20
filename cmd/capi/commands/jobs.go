package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewJobsCommand creates the jobs command group
func NewJobsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"job"},
		Short:   "Manage asynchronous jobs",
		Long:    "Monitor and manage Cloud Foundry asynchronous jobs",
	}

	cmd.AddCommand(newJobsGetCommand())
	cmd.AddCommand(newJobsPollCommand())

	return cmd
}

func newJobsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get JOB_GUID",
		Short: "Get job details",
		Long:  "Display detailed information about a specific job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			job, err := client.Jobs().Get(ctx, jobGUID)
			if err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(job)
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				return encoder.Encode(job)
			default:
				fmt.Printf("Job: %s\n", job.GUID)
				fmt.Printf("  Operation: %s\n", job.Operation)
				fmt.Printf("  State:     %s\n", job.State)
				fmt.Printf("  Created:   %s\n", job.CreatedAt.Format("2006-01-02 15:04:05"))
				fmt.Printf("  Updated:   %s\n", job.UpdatedAt.Format("2006-01-02 15:04:05"))

				if len(job.Errors) > 0 {
					fmt.Printf("  Errors:\n")
					for _, apiErr := range job.Errors {
						fmt.Printf("    - %s: %s\n", apiErr.Title, apiErr.Detail)
					}
				}

				if len(job.Warnings) > 0 {
					fmt.Printf("  Warnings:\n")
					for _, warning := range job.Warnings {
						fmt.Printf("    - %s\n", warning.Detail)
					}
				}
			}

			return nil
		},
	}
}

func newJobsPollCommand() *cobra.Command {
	var (
		interval time.Duration
		timeout  time.Duration
	)

	cmd := &cobra.Command{
		Use:   "poll JOB_GUID",
		Short: "Poll job until completion",
		Long:  "Poll a job until it completes, fails, or times out",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobGUID := args[0]

			client, err := createClient()
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Use the client's built-in polling method if available
			job, err := client.Jobs().PollUntilComplete(ctx, jobGUID)
			if err != nil {
				return fmt.Errorf("failed to poll job: %w", err)
			}

			fmt.Printf("Job completed: %s\n", job.GUID)
			fmt.Printf("  Operation: %s\n", job.Operation)
			fmt.Printf("  State:     %s\n", job.State)

			if len(job.Errors) > 0 {
				fmt.Printf("  Errors:\n")
				for _, apiErr := range job.Errors {
					fmt.Printf("    - %s: %s\n", apiErr.Title, apiErr.Detail)
				}
				return fmt.Errorf("job completed with errors")
			}

			if len(job.Warnings) > 0 {
				fmt.Printf("  Warnings:\n")
				for _, warning := range job.Warnings {
					fmt.Printf("    - %s\n", warning.Detail)
				}
			}

			return nil
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "polling interval")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "polling timeout")

	return cmd
}
