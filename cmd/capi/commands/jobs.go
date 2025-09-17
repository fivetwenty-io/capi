package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fivetwenty-io/capi/v3/internal/constants"
	"github.com/fivetwenty-io/capi/v3/pkg/capi"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// NewJobsCommand creates the jobs command group.
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

func runJobsGet(cmd *cobra.Command, args []string) error {
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

	return renderJob(job)
}

func renderJob(job *capi.Job) error {
	output := viper.GetString("output")
	switch output {
	case OutputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		err := encoder.Encode(job)
		if err != nil {
			return fmt.Errorf("failed to encode job as JSON: %w", err)
		}

		return nil
	case OutputFormatYAML:
		encoder := yaml.NewEncoder(os.Stdout)

		err := encoder.Encode(job)
		if err != nil {
			return fmt.Errorf("failed to encode job as YAML: %w", err)
		}

		return nil
	default:
		return renderJobTable(job)
	}
}

func renderJobTable(job *capi.Job) error {
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

	err := table.Render()
	if err != nil {
		return fmt.Errorf("failed to render job table: %w", err)
	}

	return nil
}

func newJobsGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get JOB_GUID",
		Short: "Get job details",
		Long:  "Display detailed information about a specific job",
		Args:  cobra.ExactArgs(1),
		RunE:  runJobsGet,
	}
}

func runJobsPoll(cmd *cobra.Command, args []string) error {
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

	err = renderJobPollResult(job)
	if err != nil {
		return err
	}

	if len(job.Errors) > 0 {
		return ErrJobCompletedWithErrors
	}

	return nil
}

func renderJobPollResult(job *capi.Job) error {
	output := viper.GetString("output")
	switch output {
	case OutputFormatJSON:
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")

		err := encoder.Encode(job)
		if err != nil {
			return fmt.Errorf("failed to encode job poll result as JSON: %w", err)
		}

		return nil
	case OutputFormatYAML:
		encoder := yaml.NewEncoder(os.Stdout)

		err := encoder.Encode(job)
		if err != nil {
			return fmt.Errorf("failed to encode job poll result as YAML: %w", err)
		}

		return nil
	default:
		_, _ = os.Stdout.WriteString("Job polling completed:\n\n")

		return renderJobTable(job)
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
		RunE:  runJobsPoll,
	}

	cmd.Flags().DurationVar(&interval, "interval", constants.DefaultJobPollInterval, "polling interval")
	cmd.Flags().DurationVar(&timeout, "timeout", constants.DefaultJobPollTimeout10, "polling timeout")

	return cmd
}
