package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler_v2"
)

func NewFrontendArtifactDataSource() datasource.DataSource {
	return &FrontendArtifactDataSource{}
}

type FrontendArtifactDataSource struct {
	client *version_handler_v2.Client
}

func (s FrontendArtifactDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_frontend_artifact"
}

func (s FrontendArtifactDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information about a specific frontend artifact version. " +
			"Artifacts are uploaded to S3 during the CI process for static website hosting. " +
			"NOTE: This data source uses the same endpoints as the Lambda Artifact data source.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource. Format: [github_repository_name]/[working_directory]",
				Computed:            true,
			},
			"github_repository_name": schema.StringAttribute{
				MarkdownDescription: "The GitHub repository name to find the artifact for.",
				Required:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "The directory in the GitHub repository to find the artifact for.",
				Optional:            true,
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
			"s3_source_path": schema.StringAttribute{
				MarkdownDescription: "The S3 source path in the format `bucket_name/object_path` where the frontend artifact is stored.",
				Computed:            true,
			},
			"s3_object_version": schema.StringAttribute{
				MarkdownDescription: "The S3 object version of the frontend artifact stored.",
				Computed:            true,
			},
			"s3_object_path": schema.StringAttribute{
				MarkdownDescription: "The S3 object path where the frontend artifact is stored.",
				Computed:            true,
			},
			"s3_bucket_name": schema.StringAttribute{
				MarkdownDescription: "The S3 bucket where the frontend artifact is stored.",
				Computed:            true,
			},
		},
	}

}

func (s *FrontendArtifactDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

type FrontendArtifactDataSourceModel struct {
	Id                   types.String `tfsdk:"id"`
	GitHubRepositoryName types.String `tfsdk:"github_repository_name"`
	WorkingDirectory     types.String `tfsdk:"working_directory"`
	GitSha               types.String `tfsdk:"git_sha"`
	Branch               types.String `tfsdk:"branch"`
	ServiceAccountID     types.String `tfsdk:"service_account_id"`
	Region               types.String `tfsdk:"region"`
	S3SourcePath         types.String `tfsdk:"s3_source_path"`
	S3ObjectPath         types.String `tfsdk:"s3_object_path"`
	S3ObjectVersion      types.String `tfsdk:"s3_object_version"`
	S3BucketName         types.String `tfsdk:"s3_bucket_name"`
}

func (s FrontendArtifactDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state FrontendArtifactDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler_v2.LambdaArtifact
	err := s.client.ReadLambdaArtifact(
		state.GitHubRepositoryName.ValueString(),
		"", // No ECR repository name for frontend artifacts
		state.WorkingDirectory.ValueString(),
		&version,
	)

	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read frontend artifact version",
			err.Error(),
		)
	}

	if workingDir := state.WorkingDirectory.ValueString(); workingDir != "" {
		state.Id = types.StringValue(fmt.Sprintf("%s/%s", state.GitHubRepositoryName.ValueString(), version.WorkingDirectory))
	} else {
		state.Id = state.GitHubRepositoryName
	}
	state.WorkingDirectory = types.StringValue(version.WorkingDirectory)
	state.GitSha = types.StringValue(version.GitSha)
	state.Branch = types.StringValue(version.Branch)
	state.ServiceAccountID = types.StringValue(version.ServiceAccountID)
	state.Region = types.StringValue(version.Region)
	state.S3ObjectPath = types.StringValue(version.S3ObjectPath)
	state.S3ObjectVersion = types.StringValue(version.S3ObjectVersion)
	state.S3BucketName = types.StringValue(version.S3BucketName)

	// Compute s3_source_path from bucket_name/object_path
	if version.S3BucketName != "" && version.S3ObjectPath != "" {
		state.S3SourcePath = types.StringValue(fmt.Sprintf("%s/%s", version.S3BucketName, version.S3ObjectPath))
	} else {
		state.S3SourcePath = types.StringNull()
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
