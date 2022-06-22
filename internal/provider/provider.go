package provider

import (
	"context"
	"fmt"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// provider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type provider struct {
	// Client can contain the upstream provider SDK or HTTP Client used to
	// communicate with the upstream service. Resource and DataSource
	// implementations can then make calls using this Client.
	Client central_cognito.Client

	// configured is set to true at the end of the Configure method.
	// This can be used in Resource and DataSource implementations to verify
	// that the provider was previously configured.
	configured bool

	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *provider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A provider for interracting with Vy's internal services.",
		Attributes: map[string]tfsdk.Attribute{
			"base_url": {
				MarkdownDescription: "The base_url for the central-cognito service",
				Type:                types.StringType,
				Optional:            true,
			},
			"environment": {
				MarkdownDescription: "The environment to provision in",
				Type:                types.StringType,
				Required:            true,
			},
		},
	}, nil
}

// providerData can be sed to store data from the Terraform configuration.
type providerData struct {
	// BaseUrl is the URL for the central-cognito service.
	BaseUrl     types.String `tfsdk:"base_url"`
	Environment types.String `tfsdk:"environment"`
}

func (p *provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var data providerData
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.BaseUrl.Null {
		data.BaseUrl.Value = "cognito.vydev.io"
		data.BaseUrl.Null = false
	}

	if data.Environment.Value == "prod" {
		p.Client.BaseUrl = fmt.Sprintf("delegated.%s", data.BaseUrl.Value)
	} else {
		p.Client.BaseUrl = fmt.Sprintf("delegated.%s.%s", data.Environment.Value, data.BaseUrl.Value)
	}

	p.configured = true
}

func (p *provider) GetResources(ctx context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"vy_resource_server": resourceServerType{},
		"vy_app_client":      appClientResourceType{},
	}, nil
}

func (p *provider) GetDataSources(ctx context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{}, nil
}

func New(version string) func() tfsdk.Provider {
	return func() tfsdk.Provider {
		return &provider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in tfsdk.Provider) (provider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*provider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return provider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return provider{}, diags
	}

	return *p, diags
}
