package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
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

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
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
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("GUID", job.GUID)
				_ = table.Append("Operation", job.Operation)
				_ = table.Append("State", job.State)
				_ = table.Append("Created", job.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", job.UpdatedAt.Format("2006-01-02 15:04:05"))

				if len(job.Errors) > 0 {
					var errorStrings []string
					for _, apiErr := range job.Errors {
						errorStrings = append(errorStrings, fmt.Sprintf("%s: %s", apiErr.Title, apiErr.Detail))
					}
					_ = table.Append("Errors", strings.Join(errorStrings, "\n"))
				}

				if len(job.Warnings) > 0 {
					var warningStrings []string
					for _, warning := range job.Warnings {
						warningStrings = append(warningStrings, warning.Detail)
					}
					_ = table.Append("Warnings", strings.Join(warningStrings, "\n"))
				}

				_ = table.Render()
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

			client, err := CreateClientWithAPI(cmd.Flag("api").Value.String())
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Use the client's built-in polling method if available
			job, err := client.Jobs().PollUntilComplete(ctx, jobGUID)
			if err != nil {
				return fmt.Errorf("failed to poll job: %w", err)
			}

			// Output results
			output := viper.GetString("output")
			switch output {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				if err := encoder.Encode(job); err != nil {
					return err
				}
			case "yaml":
				encoder := yaml.NewEncoder(os.Stdout)
				if err := encoder.Encode(job); err != nil {
					return err
				}
			default:
				table := tablewriter.NewWriter(os.Stdout)
				table.Header("Property", "Value")

				_ = table.Append("GUID", job.GUID)
				_ = table.Append("Operation", job.Operation)
				_ = table.Append("State", job.State)
				_ = table.Append("Created", job.CreatedAt.Format("2006-01-02 15:04:05"))
				_ = table.Append("Updated", job.UpdatedAt.Format("2006-01-02 15:04:05"))

				if len(job.Errors) > 0 {
					var errorStrings []string
					for _, apiErr := range job.Errors {
						errorStrings = append(errorStrings, fmt.Sprintf("%s: %s", apiErr.Title, apiErr.Detail))
					}
					_ = table.Append("Errors", strings.Join(errorStrings, "\n"))
				}

				if len(job.Warnings) > 0 {
					var warningStrings []string
					for _, warning := range job.Warnings {
						warningStrings = append(warningStrings, warning.Detail)
					}
					_ = table.Append("Warnings", strings.Join(warningStrings, "\n"))
				}

				fmt.Printf("Job polling completed:\n\n")
				_ = table.Render()
			}

			if len(job.Errors) > 0 {
				return fmt.Errorf("job completed with errors")
			}

			return nil
		},
	}

	cmd.Flags().DurationVar(&interval, "interval", 5*time.Second, "polling interval")
	cmd.Flags().DurationVar(&timeout, "timeout", 10*time.Minute, "polling timeout")

	return cmd
}
