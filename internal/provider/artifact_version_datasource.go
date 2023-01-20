package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler"
)

func NewArtifactVersionDataSource() datasource.DataSource {
	return &ArtifactVersionDataSource{}
}

type ArtifactVersionDataSource struct {
	client *version_handler.Client
}

type ArtifactVersionDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Application types.String `tfsdk:"application"`
	URI         types.String `tfsdk:"uri"`
	Store       types.String `tfsdk:"store"`
	Path        types.String `tfsdk:"path"`
	Version     types.String `tfsdk:"version"`
}

func (a ArtifactVersionDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_artifact_version"
}

func (a ArtifactVersionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "A version for an artifact",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"application": schema.StringAttribute{
				MarkdownDescription: "The application you want to find an artifact for",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The URI of the given resource",
				Computed:            true,
			},
			"store": schema.StringAttribute{
				MarkdownDescription: "The base location of where the artifact is stored",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path in the store where your application is stored",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the artifact",
				Computed:            true,
			},
		},
	}
}

func (a *ArtifactVersionDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

	a.client = configuration.VersionHandlerClient
}

func (a ArtifactVersionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state ArtifactVersionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler.Version
	err := a.client.ReadVersion(state.Application.ValueString(), &version)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read artifact version",
			fmt.Sprintf("Could not read the version for artifact %s. %s", state.Application.String(), err.Error()),
		)
	}

	state.Id = state.Application
	state.URI = types.StringValue(version.URI)
	state.Store = types.StringValue(version.Store)
	state.Path = types.StringValue(version.Path)
	state.Version = types.StringValue(version.Version)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
