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
	Id        types.String `tfsdk:"id"`
	AuthUrl   types.String `tfsdk:"auth_url"`
	JwksUrl   types.String `tfsdk:"jwks_url"`
	OpenIdUrl types.String `tfsdk:"open_id_url"`
	Issuer    types.String `tfsdk:"issuer"`
}

func (c *CognitoInfoDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_cognito_info"
}

func (c *CognitoInfoDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
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

func (c *CognitoInfoDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// TODO: This should be fetched from the service.
	//	     Doing it quickly now to get the feature shipped.
	var state CognitoInfoDataSourceModel

	if c.environment == "prod" {
		state = CognitoInfoDataSourceModel{
			AuthUrl: types.StringValue(
				"https://auth.cognito.vydev.io",
			),
			JwksUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_e6o46c1oE/.well-known/jwks.json",
			),
			OpenIdUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_e6o46c1oE/.well-known/openid-configuration",
			),
			Issuer: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_e6o46c1oE",
			),
		}
	} else if c.environment == "stage" {
		state = CognitoInfoDataSourceModel{
			AuthUrl: types.StringValue(
				"https://auth.stage.cognito.vydev.io",
			),
			JwksUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_AUYQ679zW/.well-known/jwks.json",
			),
			OpenIdUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_AUYQ679zW/.well-known/openid-configuration",
			),
			Issuer: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_AUYQ679zW",
			),
		}
	} else if c.environment == "test" {
		state = CognitoInfoDataSourceModel{
			AuthUrl: types.StringValue(
				"https://auth.test.cognito.vydev.io",
			),
			JwksUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_Z53b9AbeT/.well-known/jwks.json",
			),
			OpenIdUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_Z53b9AbeT/.well-known/openid-configuration",
			),
			Issuer: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_Z53b9AbeT",
			),
		}
	} else if c.environment == "dev" {
		state = CognitoInfoDataSourceModel{
			AuthUrl: types.StringValue(
				"https://auth.dev.cognito.vydev.io",
			),
			JwksUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_0AvVv5Wyk/.well-known/jwks.json",
			),
			OpenIdUrl: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_0AvVv5Wyk/.well-known/openid-configuration",
			),
			Issuer: types.StringValue(
				"https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_0AvVv5Wyk",
			),
		}
	}

	state.Id = types.StringValue(c.environment)

	response.State.Set(ctx, &state)
}
