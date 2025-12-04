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
	Id                   types.String `tfsdk:"id"`
	GitHubRepositoryName types.String `tfsdk:"github_repository_name"`
	WorkingDirectory     types.String `tfsdk:"working_directory"`
	GitSha               types.String `tfsdk:"git_sha"`
	Branch               types.String `tfsdk:"branch"`
	ServiceAccountID     types.String `tfsdk:"service_account_id"`
	Region               types.String `tfsdk:"region"`
	ECRRepositoryName    types.String `tfsdk:"ecr_repository_name"`
	ECRRepositoryURI     types.String `tfsdk:"ecr_repository_uri"`
}

func (e ECSImageDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest,
	response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_ecs_image"
}

func (e ECSImageDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information about a specific ECS image version. " +
			"Images are uploaded to ECR during the CI process.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource. Format: [github_repository_name]/[working_directory]",
				Computed:            true,
			},
			"github_repository_name": schema.StringAttribute{
				MarkdownDescription: "The GitHub repository name for the ECS service.",
				Required:            true,
			},
			"ecr_repository_name": schema.StringAttribute{
				MarkdownDescription: "The ECR repository name where the image to the ECS service is stored.",
				Required:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "The directory in the GitHub repository where the code is stored.",
				Optional:            true,
			},
			"git_sha": schema.StringAttribute{
				MarkdownDescription: "The Git SHA of the commit that was used to build the image.",
				Computed:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The Git branch of the commit that was used to build the image.",
				Computed:            true,
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "The service account ID that was used to build the image.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The AWS region where the image is stored.",
				Computed:            true,
			},
			"ecr_repository_uri": schema.StringAttribute{
				MarkdownDescription: "The ECR repository URI where the image is stored.",
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
	err := e.client.ReadECSImage(
		state.GitHubRepositoryName.ValueString(),
		state.ECRRepositoryName.ValueString(),
		state.WorkingDirectory.ValueString(),
		&version,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read the ECS Image",
			err.Error(),
		)
	}

	if workingDir := state.WorkingDirectory.ValueString(); workingDir != "" {
		state.Id = types.StringValue(fmt.Sprintf("%s/%s", state.GitHubRepositoryName.ValueString(), workingDir))
	} else {
		state.Id = state.GitHubRepositoryName
	}
	state.WorkingDirectory = types.StringValue(version.WorkingDirectory)
	state.GitSha = types.StringValue(version.GitSha)
	state.Branch = types.StringValue(version.Branch)
	state.ServiceAccountID = types.StringValue(version.ServiceAccountID)
	state.Region = types.StringValue(version.Region)

	// If overrides the repo name
	if !state.ECRRepositoryName.IsNull() && state.ECRRepositoryName.ValueString() != "" {
		state.ECRRepositoryName = types.StringValue(state.ECRRepositoryName.ValueString())
		state.ECRRepositoryURI = types.StringValue(fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s",
			state.ServiceAccountID.ValueString(), state.Region.ValueString(), state.ECRRepositoryName.ValueString()))
	} else {
		// Use values from API
		state.ECRRepositoryName = types.StringValue(version.ECRRepositoryName)
		state.ECRRepositoryURI = types.StringValue(version.ECRRepositoryURI)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
