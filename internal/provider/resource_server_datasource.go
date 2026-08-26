package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

var _ datasource.DataSource = &ResourceServerDataSource{}

func NewResourceServerDataSource() datasource.DataSource {
	return &ResourceServerDataSource{}
}

type ResourceServerDataSource struct {
	client *central_cognito.Client
}

type ResourceServerDataSourceModel struct {
	Id         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Scopes     []scope      `tfsdk:"scopes"`
}

func (d *ResourceServerDataSource) Metadata(ctx context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource_server"
}

func (d *ResourceServerDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Looks up an existing resource server by its identifier. Pair this with " +
			"`vy_resource_server_scope` to add scopes to a resource server owned by another module or " +
			"team, without needing to manage the resource server itself.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of this resource. Same value as `identifier`.",
			},
			"identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The identity of the resource server to look up.",
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of this resource server.",
			},
			"scopes": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The scopes currently registered on this resource server.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of this scope.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "A description of what this scope is for.",
						},
					},
				},
			},
		},
	}
}

func (d *ResourceServerDataSource) Configure(ctx context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

	d.client = configuration.CognitoClient
}

func (d *ResourceServerDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data ResourceServerDataSourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var server central_cognito.ResourceServer
	if err := d.client.ReadResourceServer(data.Identifier.ValueString(), &server); err != nil {
		response.Diagnostics.AddError(
			"Unable to read resource server",
			fmt.Sprintf("Could not read resource server %q: %s", data.Identifier.ValueString(), err.Error()),
		)
		return
	}

	data.Id = types.StringValue(server.Identifier)
	data.Identifier = types.StringValue(server.Identifier)
	data.Name = types.StringValue(server.Name)

	data.Scopes = []scope{}
	for _, s := range server.Scopes {
		data.Scopes = append(data.Scopes, scope{
			Name:        types.StringValue(s.Name),
			Description: types.StringValue(s.Description),
		})
	}

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}
