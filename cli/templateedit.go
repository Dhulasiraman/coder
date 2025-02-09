package cli

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/xerrors"

	"github.com/coder/coder/cli/cliui"
	"github.com/coder/coder/codersdk"
)

func templateEdit() *cobra.Command {
	var (
		name                         string
		displayName                  string
		description                  string
		icon                         string
		defaultTTL                   time.Duration
		maxTTL                       time.Duration
		allowUserCancelWorkspaceJobs bool
	)

	cmd := &cobra.Command{
		Use:   "edit <template> [flags]",
		Args:  cobra.ExactArgs(1),
		Short: "Edit the metadata of a template by name.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := CreateClient(cmd)
			if err != nil {
				return xerrors.Errorf("create client: %w", err)
			}

			if maxTTL != 0 {
				entitlements, err := client.Entitlements(cmd.Context())
				var sdkErr *codersdk.Error
				if xerrors.As(err, &sdkErr) && sdkErr.StatusCode() == http.StatusNotFound {
					return xerrors.Errorf("your deployment appears to be an AGPL deployment, so you cannot set --max-ttl")
				} else if err != nil {
					return xerrors.Errorf("get entitlements: %w", err)
				}

				if !entitlements.Features[codersdk.FeatureAdvancedTemplateScheduling].Enabled {
					return xerrors.Errorf("your license is not entitled to use advanced template scheduling, so you cannot set --max-ttl")
				}
			}

			organization, err := CurrentOrganization(cmd, client)
			if err != nil {
				return xerrors.Errorf("get current organization: %w", err)
			}
			template, err := client.TemplateByName(cmd.Context(), organization.ID, args[0])
			if err != nil {
				return xerrors.Errorf("get workspace template: %w", err)
			}

			// NOTE: coderd will ignore empty fields.
			req := codersdk.UpdateTemplateMeta{
				Name:                         name,
				DisplayName:                  displayName,
				Description:                  description,
				Icon:                         icon,
				DefaultTTLMillis:             defaultTTL.Milliseconds(),
				MaxTTLMillis:                 maxTTL.Milliseconds(),
				AllowUserCancelWorkspaceJobs: allowUserCancelWorkspaceJobs,
			}

			_, err = client.UpdateTemplateMeta(cmd.Context(), template.ID, req)
			if err != nil {
				return xerrors.Errorf("update template metadata: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated template metadata at %s!\n", cliui.Styles.DateTimeStamp.Render(time.Now().Format(time.Stamp)))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "", "", "Edit the template name.")
	cmd.Flags().StringVarP(&displayName, "display-name", "", "", "Edit the template display name.")
	cmd.Flags().StringVarP(&description, "description", "", "", "Edit the template description.")
	cmd.Flags().StringVarP(&icon, "icon", "", "", "Edit the template icon path.")
	cmd.Flags().DurationVarP(&defaultTTL, "default-ttl", "", 0, "Edit the template default time before shutdown - workspaces created from this template default to this value.")
	cmd.Flags().DurationVarP(&maxTTL, "max-ttl", "", 0, "Edit the template maximum time before shutdown - workspaces created from this template must shutdown within the given duration after starting. This is an enterprise-only feature.")
	cmd.Flags().BoolVarP(&allowUserCancelWorkspaceJobs, "allow-user-cancel-workspace-jobs", "", true, "Allow users to cancel in-progress workspace jobs.")
	cliui.AllowSkipPrompt(cmd)

	return cmd
}
