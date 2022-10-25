package codersdk

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// TemplateVersion represents a single version of a template.
type TemplateVersion struct {
	ID             uuid.UUID      `json:"id"`
	TemplateID     *uuid.UUID     `json:"template_id,omitempty"`
	OrganizationID uuid.UUID      `json:"organization_id,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	Name           string         `json:"name"`
	Job            ProvisionerJob `json:"job"`
	Readme         string         `json:"readme"`
	CreatedByID    uuid.UUID      `json:"created_by_id"`
	CreatedByName  string         `json:"created_by_name"`
}

// TemplateVersionParameter represents a parameter for a template version.
type TemplateVersionParameter struct {
	Name            string                           `json:"name"`
	Description     string                           `json:"description"`
	Type            string                           `json:"type"`
	Mutable         bool                             `json:"mutable"`
	DefaultValue    string                           `json:"default_value"`
	Icon            string                           `json:"icon"`
	Options         []TemplateVersionParameterOption `json:"options"`
	ValidationError string                           `json:"validation_error"`
	ValidationRegex string                           `json:"validation_regex"`
	ValidationMin   int32                            `json:"validation_min"`
	ValidationMax   int32                            `json:"validation_max"`
}

// TemplateVersionParameterOption represents a selectable option for a template parameter.
type TemplateVersionParameterOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Icon        string `json:"icon"`
}

// TemplateVersion returns a template version by ID.
func (c *Client) TemplateVersion(ctx context.Context, id uuid.UUID) (TemplateVersion, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s", id), nil)
	if err != nil {
		return TemplateVersion{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return TemplateVersion{}, readBodyAsError(res)
	}
	var version TemplateVersion
	return version, json.NewDecoder(res.Body).Decode(&version)
}

// CancelTemplateVersion marks a template version job as canceled.
func (c *Client) CancelTemplateVersion(ctx context.Context, version uuid.UUID) error {
	res, err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("/api/v2/templateversions/%s/cancel", version), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return readBodyAsError(res)
	}
	return nil
}

// TemplateVersionParameters returns parameters a template version exposes.
func (c *Client) TemplateVersionParameters(ctx context.Context, version uuid.UUID) ([]TemplateVersionParameter, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/parameters", version), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}
	var params []TemplateVersionParameter
	return params, json.NewDecoder(res.Body).Decode(&params)
}

// DeprecatedTemplateVersionSchema returns schemas for a template version by ID.
func (c *Client) DeprecatedTemplateVersionSchema(ctx context.Context, version uuid.UUID) ([]DeprecatedParameterSchema, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/deprecated-schema", version), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}
	var params []DeprecatedParameterSchema
	return params, json.NewDecoder(res.Body).Decode(&params)
}

// DeprecatedTemplateVersionParameters returns computed parameters for a template version.
func (c *Client) DeprecatedTemplateVersionParameters(ctx context.Context, version uuid.UUID) ([]DeprecatedComputedParameter, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/deprecated-parameters", version), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}
	var params []DeprecatedComputedParameter
	return params, json.NewDecoder(res.Body).Decode(&params)
}

// TemplateVersionResources returns resources a template version declares.
func (c *Client) TemplateVersionResources(ctx context.Context, version uuid.UUID) ([]WorkspaceResource, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/resources", version), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}
	var resources []WorkspaceResource
	return resources, json.NewDecoder(res.Body).Decode(&resources)
}

// TemplateVersionLogsBefore returns logs that occurred before a specific time.
func (c *Client) TemplateVersionLogsBefore(ctx context.Context, version uuid.UUID, before time.Time) ([]ProvisionerJobLog, error) {
	return c.provisionerJobLogsBefore(ctx, fmt.Sprintf("/api/v2/templateversions/%s/logs", version), before)
}

// TemplateVersionLogsAfter streams logs for a template version that occurred after a specific time.
func (c *Client) TemplateVersionLogsAfter(ctx context.Context, version uuid.UUID, after time.Time) (<-chan ProvisionerJobLog, io.Closer, error) {
	return c.provisionerJobLogsAfter(ctx, fmt.Sprintf("/api/v2/templateversions/%s/logs", version), after)
}

// CreateTemplateVersionDryRunRequest defines the request parameters for
// CreateTemplateVersionDryRun.
type CreateTemplateVersionDryRunRequest struct {
	WorkspaceName   string                             `json:"workspace_name"`
	ParameterValues []DeprecatedCreateParameterRequest `json:"parameter_values"`
}

// CreateTemplateVersionDryRun begins a dry-run provisioner job against the
// given template version with the given parameter values.
func (c *Client) CreateTemplateVersionDryRun(ctx context.Context, version uuid.UUID, req CreateTemplateVersionDryRunRequest) (ProvisionerJob, error) {
	res, err := c.Request(ctx, http.MethodPost, fmt.Sprintf("/api/v2/templateversions/%s/dry-run", version), req)
	if err != nil {
		return ProvisionerJob{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusCreated {
		return ProvisionerJob{}, readBodyAsError(res)
	}

	var job ProvisionerJob
	return job, json.NewDecoder(res.Body).Decode(&job)
}

// TemplateVersionDryRun returns the current state of a template version dry-run
// job.
func (c *Client) TemplateVersionDryRun(ctx context.Context, version, job uuid.UUID) (ProvisionerJob, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/dry-run/%s", version, job), nil)
	if err != nil {
		return ProvisionerJob{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ProvisionerJob{}, readBodyAsError(res)
	}

	var j ProvisionerJob
	return j, json.NewDecoder(res.Body).Decode(&j)
}

// TemplateVersionDryRunResources returns the resources of a finished template
// version dry-run job.
func (c *Client) TemplateVersionDryRunResources(ctx context.Context, version, job uuid.UUID) ([]WorkspaceResource, error) {
	res, err := c.Request(ctx, http.MethodGet, fmt.Sprintf("/api/v2/templateversions/%s/dry-run/%s/resources", version, job), nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, readBodyAsError(res)
	}

	var resources []WorkspaceResource
	return resources, json.NewDecoder(res.Body).Decode(&resources)
}

// TemplateVersionDryRunLogsBefore returns logs for a template version dry-run
// that occurred before a specific time.
func (c *Client) TemplateVersionDryRunLogsBefore(ctx context.Context, version, job uuid.UUID, before time.Time) ([]ProvisionerJobLog, error) {
	return c.provisionerJobLogsBefore(ctx, fmt.Sprintf("/api/v2/templateversions/%s/dry-run/%s/logs", version, job), before)
}

// TemplateVersionDryRunLogsAfter streams logs for a template version dry-run
// that occurred after a specific time.
func (c *Client) TemplateVersionDryRunLogsAfter(ctx context.Context, version, job uuid.UUID, after time.Time) (<-chan ProvisionerJobLog, io.Closer, error) {
	return c.provisionerJobLogsAfter(ctx, fmt.Sprintf("/api/v2/templateversions/%s/dry-run/%s/logs", version, job), after)
}

// CancelTemplateVersionDryRun marks a template version dry-run job as canceled.
func (c *Client) CancelTemplateVersionDryRun(ctx context.Context, version, job uuid.UUID) error {
	res, err := c.Request(ctx, http.MethodPatch, fmt.Sprintf("/api/v2/templateversions/%s/dry-run/%s/cancel", version, job), nil)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return readBodyAsError(res)
	}
	return nil
}
