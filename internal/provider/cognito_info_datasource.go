package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type cognitoInfoDatasourceType struct{}

func (c cognitoInfoDatasourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"auth_url": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The URL where users can authenticate",
			},
			"jwks_url": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The URL for the /.well-known/jwks.json",
			},
			"open_id_url": {
				Type:                types.StringType,
				Computed:            true,
				MarkdownDescription: "The URL for the /.well-known/openid-configuration",
			},
		},
	}, nil
}

func (c cognitoInfoDatasourceType) NewDataSource(_ context.Context, p tfsdk.Provider) (tfsdk.DataSource, diag.Diagnostics) {
	provider, diags := convertProviderType(p)

	return cognitoInfo{
		provider: provider,
	}, diags
}

type cognitoInfo struct {
	provider provider
}

type cognitoInfoData struct {
	Id        types.String `tfsdk:"id"`
	AuthUrl   types.String `tfsdk:"auth_url"`
	JwksUrl   types.String `tfsdk:"jwks_url"`
	OpenIdUrl types.String `tfsdk:"open_id_url"`
}

func (c cognitoInfo) Read(ctx context.Context, request tfsdk.ReadDataSourceRequest, response *tfsdk.ReadDataSourceResponse) {
	// TODO: This should be fetched from the service.
	//	     Doing it quickly now to get the feature shipped.
	var state cognitoInfoData

	state.Id.Value = c.provider.CentralCognitoEnvironment

	if c.provider.CentralCognitoEnvironment == "prod" {
		state = cognitoInfoData{
			AuthUrl: types.String{
				Value: "https://auth.cognito.vydev.io",
			},
			JwksUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_e6o46c1oE/.well-known/jwks.json",
			},
			OpenIdUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_e6o46c1oE/.well-known/openid-configuration",
			},
		}
	} else if c.provider.CentralCognitoEnvironment == "stage" {
		state = cognitoInfoData{
			AuthUrl: types.String{
				Value: "https://auth.stage.cognito.vydev.io",
			},
			JwksUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_AUYQ679zW/.well-known/jwks.json",
			},
			OpenIdUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_AUYQ679zW/.well-known/openid-configuration",
			},
		}
	} else if c.provider.CentralCognitoEnvironment == "test" {
		state = cognitoInfoData{
			AuthUrl: types.String{
				Value: "https://auth.test.cognito.vydev.io",
			},
			JwksUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_Z53b9AbeT/.well-known/jwks.json",
			},
			OpenIdUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_Z53b9AbeT/.well-known/openid-configuration",
			},
		}
	} else if c.provider.CentralCognitoEnvironment == "dev" {
		state = cognitoInfoData{
			AuthUrl: types.String{
				Value: "https://auth.dev.cognito.vydev.io",
			},
			JwksUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_0AvVv5Wyk/.well-known/jwks.json",
			},
			OpenIdUrl: types.String{
				Value: "https://cognito-idp.eu-west-1.amazonaws.com/eu-west-1_0AvVv5Wyk/.well-known/openid-configuration",
			},
		}
	}

	response.State.Set(ctx, &state)
}
