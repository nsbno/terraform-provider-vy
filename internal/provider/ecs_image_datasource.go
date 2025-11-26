package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler_v2"
)

func NewECSImageDataSource() datasource.DataSource {
	return &ECSImageDataSource{}
}

type ECSImageDataSource struct {
	client *version_handler_v2.Client
}

type ECSImageDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	ECRRepositoryName types.String `tfsdk:"ecr_repository_name"`
	URI               types.String `tfsdk:"uri"`
	Store             types.String `tfsdk:"store"`
	Path              types.String `tfsdk:"path"`
	Version           types.String `tfsdk:"version"`
	GitSha            types.String `tfsdk:"git_sha"`
}

func (e ECSImageDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest,
	response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_ecs_image"
}

func (e ECSImageDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information about a specific artifact version. " +
			"Artifacts are uploaded to ECR during the CI process. " +
			"Each unique service (Lambda or ECS) should have its own ECR repository.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"ecr_repository_name": schema.StringAttribute{
				MarkdownDescription: "The ECR repository name to find the image for.",
				Required:            true,
			},
			"uri": schema.StringAttribute{
				MarkdownDescription: "The full Image URI of the ECR Image. Prefixed with docker://",
				Computed:            true,
			},
			"store": schema.StringAttribute{
				MarkdownDescription: "The ECR URI, in this format: `{registry_id}.dkr.ecr.{region}.amazonaws.com`",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The ECR repository name where the image is stored.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the ECR Image, which is the image digest.",
				Computed:            true,
			},
			"git_sha": schema.StringAttribute{
				MarkdownDescription: "The Git SHA of the commit that was used to build the image. Used to tag the image.",
				Computed:            true,
			},
		},
	}
}

func (e *ECSImageDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func (e ECSImageDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state ECSImageDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler_v2.ECSVersion
	err := e.client.ReadECSImage(state.ECRRepositoryName.ValueString(), &version)
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
