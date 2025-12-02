package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler_v2"
)

func NewLambdaArtifactDataSource() datasource.DataSource {
	return &LambdaArtifactDataSource{}
}

type LambdaArtifactDataSource struct {
	client *version_handler_v2.Client
}

func (s LambdaArtifactDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_lambda_artifact"
}

func (s LambdaArtifactDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information about a specific Lambda artifact version. " +
			"Artifacts are uploaded to S3 or ECR during the CI process.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"github_repository_name": schema.StringAttribute{
				MarkdownDescription: "The GitHub repository name to find the artifact for.",
				Required:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "The directory in the GitHub repository to find the artifact for.",
				Optional:            true,
			},
			"ecr_repository_name": schema.StringAttribute{
				MarkdownDescription: "*Only if artifact type is ECR.* " +
					"The ECR repository name where the Lambda image is stored.",
				Optional: true,
			},
			"git_sha": schema.StringAttribute{
				MarkdownDescription: "The Git SHA of the commit that was used to build the artifact.",
				Computed:            true,
			},
			"branch": schema.StringAttribute{
				MarkdownDescription: "The Git branch of the commit that was used to build the artifact.",
				Computed:            true,
			},
			"service_account_id": schema.StringAttribute{
				MarkdownDescription: "The service account ID that was used to build the artifact.",
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The AWS region where the artifact is stored.",
				Computed:            true,
			},
			"bucket_name": schema.StringAttribute{
				MarkdownDescription: "*Only if artifact type is S3.* The S3 bucket name where the Lambda artifact is stored.",
				Computed:            true,
			},
		},
	}

}

func (s *LambdaArtifactDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

	s.client = configuration.VersionHandlerClientV2
}

type LambdaArtifactDataSourceModel struct {
	Id                   types.String `tfsdk:"id"`
	GitHubRepositoryName types.String `tfsdk:"github_repository_name"`
	WorkingDirectory     types.String `tfsdk:"working_directory"`
	GitSha               types.String `tfsdk:"git_sha"`
	Branch               types.String `tfsdk:"branch"`
	ServiceAccountID     types.String `tfsdk:"service_account_id"`
	ECRRepositoryName    types.String `tfsdk:"ecr_repository_name"`
	Region               types.String `tfsdk:"region"`
	BucketName           types.String `tfsdk:"bucket_name"`
}

func (s LambdaArtifactDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state LambdaArtifactDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler_v2.LambdaArtifact
	err := s.client.ReadLambdaArtifact(
		state.GitHubRepositoryName.ValueString(),
		state.ECRRepositoryName.ValueString(),
		state.WorkingDirectory.ValueString(),
		&version,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read Lambda artifact version",
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
	state.ECRRepositoryName = types.StringValue(version.ECRRepositoryName)
	state.BucketName = types.StringValue(version.BucketName)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
