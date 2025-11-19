package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/version_handler_v2"
)

func NewS3ArtifactDataSource() datasource.DataSource {
	return &S3ArtifactDataSource{}
}

type S3ArtifactDataSource struct {
	client *version_handler_v2.Client
}

func (s S3ArtifactDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_s3_artifact"
}

func (s S3ArtifactDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Get information from a specific artifact version in S3. " +
			"Artifacts are uploaded to S3 during the CI process. " +
			"We assume that each GitHub Repository in a given directory has the same artifact version based on" +
			" Git sha. e.g. multiple lambda functions in a directory in a GitHub repository will have the same" +
			" artifact version.",

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
			"uri": schema.StringAttribute{
				MarkdownDescription: "The URI of the S3 artifact.",
				Computed:            true,
			},
			"store": schema.StringAttribute{
				MarkdownDescription: "The S3 Bucket name where the artifact is stored.",
				Computed:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "The S3 key for the artifact.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The version of the S3 artifact, which is the object version ID.",
				Computed:            true,
			},
			"git_sha": schema.StringAttribute{
				MarkdownDescription: "The Git SHA of the commit that was used to build the artifact. " +
					"Used as S3 filename.",
				Computed: true,
			},
		},
	}

}

func (s *S3ArtifactDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

type S3ArtifactDataSourceModel struct {
	Id                   types.String `tfsdk:"id"`
	GitHubRepositoryName types.String `tfsdk:"github_repository_name"`
	WorkingDirectory     types.String `tfsdk:"working_directory"`
	URI                  types.String `tfsdk:"uri"`
	Store                types.String `tfsdk:"store"`
	Path                 types.String `tfsdk:"path"`
	Version              types.String `tfsdk:"version"`
	GitSha               types.String `tfsdk:"git_sha"`
}

func (s S3ArtifactDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state S3ArtifactDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	var version version_handler_v2.S3Artifact
	err := s.client.ReadS3Artifact(state.GitHubRepositoryName.ValueString(), state.WorkingDirectory.ValueString(), &version)
	if err != nil {
		errorMessage := fmt.Sprintf("Could not find the version for artifact %s. %s",
			state.GitHubRepositoryName.String(), err.Error())
		if workingDir := state.WorkingDirectory.ValueString(); workingDir != "" {
			errorMessage = fmt.Sprintf("Could not find the version for artifact %s/%s. %s", state.GitHubRepositoryName.String(), workingDir, err.Error())
		}
		response.Diagnostics.AddError(
			"Unable to read S3 artifact version",
			errorMessage,
		)
	}

	if workingDir := state.WorkingDirectory.ValueString(); workingDir != "" {
		state.Id = types.StringValue(fmt.Sprintf("%s/%s", state.GitHubRepositoryName.ValueString(), workingDir))
	} else {
		state.Id = state.GitHubRepositoryName
	}
	state.URI = types.StringValue(version.URI)
	state.Store = types.StringValue(version.Store)
	state.Path = types.StringValue(version.Path)
	state.Version = types.StringValue(version.Version)
	state.GitSha = types.StringValue(version.GitSha)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
