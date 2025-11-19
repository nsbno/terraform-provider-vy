package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler_v2"
)

func NewECRImageDataSource() datasource.DataSource {
	return &ECRImageDataSource{}
}

type ECRImageDataSource struct {
	client *version_handler_v2.Client
}

type ECRImageDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	ECRRepositoryName types.String `tfsdk:"ecr_repository_name"`
	URI               types.String `tfsdk:"uri"`
	Store             types.String `tfsdk:"store"`
	Path              types.String `tfsdk:"path"`
	Version           types.String `tfsdk:"version"`
	GitSha            types.String `tfsdk:"git_sha"`
}

func (e ECRImageDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest,
	response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_ecr_image"
}

func (e ECRImageDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information about a specific artifact version. " +
			"Artifacts are uploaded to ECR during the CI process.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ecr_repository_name": schema.StringAttribute{
				MarkdownDescription: "The ECR repository name to find the image for.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The Image URI of the ECR Image.",
				Computed:            true,
			},
			"store": schema.StringAttribute{
				MarkdownDescription: "The base location of where the artifact is stored. ECR.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The path in ECR where your image is stored, which is the image tag.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the ECR Image, which is the image digest.",
				Computed:            true,
			},
			"git_sha": schema.StringAttribute{
				MarkdownDescription: "The Git SHA of the commit that was used to build the image.",
				Computed:            true,
			},
		},
	}
}

func (e *ECRImageDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

	e.client = configuration.VersionHandlerClientV2
}

func (e ECRImageDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state ECRImageDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler_v2.ECRVersion
	err := e.client.ReadECRImage(state.ECRRepositoryName.ValueString(), &version)
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to find the ECR Image",
			fmt.Sprintf("Could not find image in ECR Repo: %s. %s", state.ECRRepositoryName.String(), err.Error()),
		)
	}

	state.Id = state.ECRRepositoryName
	state.URI = types.StringValue(version.URI)
	state.Store = types.StringValue(version.Store)
	state.Path = types.StringValue(version.Path)
	state.Version = types.StringValue(version.Version)
	state.GitSha = types.StringValue(version.GitSha)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
