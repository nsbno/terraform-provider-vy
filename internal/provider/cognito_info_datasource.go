package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

var _ datasource.DataSource = &CognitoInfoDataSource{}

func NewCognitoInfoDataSource() datasource.DataSource {
	return &CognitoInfoDataSource{}
}

type CognitoInfoDataSource struct {
	environment string
	client      *central_cognito.Client
}

type CognitoInfoDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Environment types.String `tfsdk:"environment"`
	UserPoolId  types.String `tfsdk:"user_pool_id"`
	UserPoolArn types.String `tfsdk:"user_pool_arn"`
	AuthUrl     types.String `tfsdk:"auth_url"`
	JwksUrl     types.String `tfsdk:"jwks_url"`
	OpenIdUrl   types.String `tfsdk:"open_id_url"`
	Issuer      types.String `tfsdk:"issuer"`
}

func (c *CognitoInfoDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cognito_info"
}

func (c *CognitoInfoDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The environment name (e.g. prod, stage, test, dev). Same value as `environment`.",
			},
			"environment": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The environment name (e.g. prod, stage, test, dev)",
			},
			"user_pool_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the Cognito User Pool",
			},
			"user_pool_arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the Cognito User Pool",
			},
			"auth_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL where users can authenticate",
			},
			"jwks_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL for the /.well-known/jwks.json",
			},
			"open_id_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URL for the /.well-known/openid-configuration",
			},
			"issuer": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The URI for the issuer",
			},
		},
	}
}

func (c *CognitoInfoDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	configuration, ok := request.ProviderData.(*VyProviderConfiguration)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *VyProviderConfiguration, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	c.environment = configuration.Environment
	c.client = configuration.CognitoClient
}

type cognitoEnvironmentConfig struct {
	userPoolId string
	accountId  string
	authUrl    string
}

var cognitoEnvironments = map[string]cognitoEnvironmentConfig{
	"prod":  {userPoolId: "eu-west-1_e6o46c1oE", accountId: "387958190215", authUrl: "https://auth.cognito.vydev.io"},
	"stage": {userPoolId: "eu-west-1_AUYQ679zW", accountId: "214014793664", authUrl: "https://auth.stage.cognito.vydev.io"},
	"test":  {userPoolId: "eu-west-1_Z53b9AbeT", accountId: "231176028624", authUrl: "https://auth.test.cognito.vydev.io"},
	"dev":   {userPoolId: "eu-west-1_0AvVv5Wyk", accountId: "834626710667", authUrl: "https://auth.dev.cognito.vydev.io"},
}

func cognitoStateFromConfig(environment string, cognitoConfig cognitoEnvironmentConfig) CognitoInfoDataSourceModel {
	cognitoBase := fmt.Sprintf("https://cognito-idp.eu-west-1.amazonaws.com/%s", cognitoConfig.userPoolId)
	return CognitoInfoDataSourceModel{
		Id:          types.StringValue(environment),
		Environment: types.StringValue(environment),
		UserPoolId:  types.StringValue(cognitoConfig.userPoolId),
		UserPoolArn: types.StringValue(fmt.Sprintf("arn:aws:cognito-idp:eu-west-1:%s:userpool/%s", cognitoConfig.accountId, cognitoConfig.userPoolId)),
		AuthUrl:     types.StringValue(cognitoConfig.authUrl),
		JwksUrl:     types.StringValue(cognitoBase + "/.well-known/jwks.json"),
		OpenIdUrl:   types.StringValue(cognitoBase + "/.well-known/openid-configuration"),
		Issuer:      types.StringValue(cognitoBase),
	}
}

func (c *CognitoInfoDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// TODO: This should be fetched from the service.
	//	     Doing it quickly now to get the feature shipped.
	cognitoConfig, found := cognitoEnvironments[c.environment]
	if !found {
		response.Diagnostics.AddError(
			"Unsupported environment",
			fmt.Sprintf("No Cognito configuration found for environment %q", c.environment),
		)
		return
	}

	state := cognitoStateFromConfig(c.environment, cognitoConfig)
	response.State.Set(ctx, &state)
}
