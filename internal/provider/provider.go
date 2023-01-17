package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
	"github.com/nsbno/terraform-provider-vy/internal/enroll_account"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &VyProvider{}

// VyProvider satisfies the tfsdk.Provider interface and usually is included
// with all Resource and DataSource implementations.
type VyProvider struct {
	// version is set to the VyProvider version on release, "dev" when the
	// VyProvider is built and ran locally, and "test" when running acceptance
	// testing.
	version string

	config *VyProviderConfiguration
}

type VyProviderConfiguration struct {
	Environment         string
	CognitoClient       *central_cognito.Client
	EnrollAccountClient *enroll_account.Client
}

// VyProviderModel can be used to store data from the Terraform configuration.
type VyProviderModel struct {
	CentralCognitoBaseUrl types.String `tfsdk:"central_cognito_base_url"`
	EnrollAccountBaseUrl  types.String `tfsdk:"enroll_account_base_url"`
	Environment           types.String `tfsdk:"environment"`
}

func (p VyProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "vy"
	response.Version = p.version
}

func (p VyProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "A VyProvider for interracting with Vy's internal services.",
		Attributes: map[string]schema.Attribute{
			"central_cognito_base_url": schema.StringAttribute{
				MarkdownDescription: "The base url for the central-cognito service",
				Optional:            true,
			},
			"enroll_account_base_url": schema.StringAttribute{
				MarkdownDescription: "The base url for the deployment enrollment service",
				Optional:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "The environment to provision in",
				Required:            true,
			},
		},
	}
}

func createUrlFromEnvironment(baseUrl string, urlPrefix string, environment string) string {
	if environment == "prod" {
		return fmt.Sprintf("%s.%s", urlPrefix, baseUrl)
	} else {
		return fmt.Sprintf("%s.%s.%s", urlPrefix, environment, baseUrl)
	}
}

func (p VyProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var data VyProviderModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	cognito_domain := "cognito.vydev.io"
	if !data.CentralCognitoBaseUrl.IsNull() {
		cognito_domain = data.CentralCognitoBaseUrl.ValueString()
	}

	deployment_domain := "vydeployment.vydev.io"
	if !data.EnrollAccountBaseUrl.IsNull() {
		deployment_domain = data.EnrollAccountBaseUrl.ValueString()
	}

	cognito_client := &central_cognito.Client{
		BaseUrl: createUrlFromEnvironment(cognito_domain, "delegated", data.Environment.ValueString()),
	}

	enroll_client := &enroll_account.Client{
		BaseUrl: createUrlFromEnvironment(deployment_domain, "enroll", data.Environment.ValueString()),
	}

	config := &VyProviderConfiguration{
		Environment:         data.Environment.ValueString(),
		CognitoClient:       cognito_client,
		EnrollAccountClient: enroll_client,
	}

	p.config = config
	response.ResourceData = config
	response.DataSourceData = config
}

func (p VyProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewResourceServerResource,
		NewAppClientResource,
		NewDeploymentResource,
	}
}

func (p VyProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCognitoInfoDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VyProvider{
			version: version,
		}
	}
}

// convertProviderType is a helper function for NewResource and NewDataSource
// implementations to associate the concrete provider type. Alternatively,
// this helper can be skipped and the provider type can be directly type
// asserted (e.g. provider: in.(*provider)), however using this can prevent
// potential panics.
func convertProviderType(in provider.Provider) (*VyProvider, diag.Diagnostics) {
	var diags diag.Diagnostics

	p, ok := in.(*VyProvider)

	if !ok {
		diags.AddError(
			"Unexpected Provider Instance Type",
			fmt.Sprintf("While creating the data source or resource, an unexpected provider type (%T) was received. This is always a bug in the provider code and should be reported to the provider developers.", p),
		)
		return &VyProvider{}, diags
	}

	if p == nil {
		diags.AddError(
			"Unexpected Provider Instance Type",
			"While creating the data source or resource, an unexpected empty provider instance was received. This is always a bug in the provider code and should be reported to the provider developers.",
		)
		return &VyProvider{}, diags
	}

	return p, diags
}
