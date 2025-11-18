package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

func NewResourceServerResource() resource.Resource {
	return &ResourceServerResource{}
}

type ResourceServerResource struct {
	client *central_cognito.Client
}

type ResourceServerResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Scopes     []scope      `tfsdk:"scopes"`
}

type scope struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

func (r ResourceServerResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource_server"
}

func (r ResourceServerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "A resource server is an integration between a user pool and an API. " +
			"Each resource server has custom scopes that you must activate in your app client. " +
			"When you configure a resource server, your app can generate access tokens with OAuth scopes that " +
			"authorize read and write operations to your API server.",

		Attributes: map[string]schema.Attribute{
			// id is required by the SDKv2 testing framework.
			// See https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The identity of this resource server",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of this resource server",
				Required:            true,
			},
			"scopes": schema.SetNestedAttribute{
				MarkdownDescription: "Scopes for this resource server",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "A name for this scope",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "A description of what this scope is for",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (c *ResourceServerResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	c.client = configuration.CognitoClient
}

func stateToDomain(state ResourceServerResourceModel, domain *central_cognito.ResourceServer) {
	domain.Identifier = state.Identifier.ValueString()
	domain.Name = state.Name.ValueString()

	domain.Scopes = []central_cognito.Scope{}

	for _, state_scope := range state.Scopes {
		domain.Scopes = append(domain.Scopes, central_cognito.Scope{
			Name:        state_scope.Name.ValueString(),
			Description: state_scope.Description.ValueString(),
		})
	}
}

func domainToState(domain central_cognito.ResourceServer, state *ResourceServerResourceModel) {
	// See schema's comment about id to see why we do this.
	state.Id = types.StringValue(domain.Identifier)
	state.Identifier = types.StringValue(domain.Identifier)
	state.Name = types.StringValue(domain.Name)

	state.Scopes = []scope{}
	for _, domain_scope := range domain.Scopes {
		state_scope := scope{}

		state_scope.Name = types.StringValue(domain_scope.Name)
		state_scope.Description = types.StringValue(domain_scope.Description)

		state.Scopes = append(state.Scopes, state_scope)
	}

	if len(state.Scopes) == 0 {
		// Terraform thinks this has changed if we keep the scopes as an empty list.
		state.Scopes = nil
	}
}

func (r ResourceServerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ResourceServerResourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Id = data.Identifier

	var server central_cognito.ResourceServer
	stateToDomain(data, &server)

	err := r.client.CreateResourceServer(server)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Could not create resource server",
			fmt.Sprintf("Resource server with ID %s could not be created: %s", server.Identifier, err.Error()),
		)
		response.Diagnostics.Append(diags...)

		return
	}

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ResourceServerResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var server central_cognito.ResourceServer
	err := r.client.ReadResourceServer(data.Identifier.ValueString(), &server)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to read resource server",
			"Can't read resource server "+data.Identifier.String()+" from remote: "+err.Error(),
		)
		response.Diagnostics.Append(diags...)
		return
	}

	domainToState(server, &data)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data ResourceServerResourceModel

	diags := request.Plan.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Id = data.Identifier

	var server central_cognito.ResourceServer
	stateToDomain(data, &server)

	err := r.client.UpdateResourceServer(central_cognito.ResourceServerUpdateRequest{
		Identifier: server.Identifier,
		Name:       server.Name,
		Scopes:     server.Scopes,
	})
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to update resource server",
			"Can't update resource server "+data.Identifier.String()+": "+err.Error(),
		)
		response.Diagnostics.Append(diags...)

		return
	}

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ResourceServerResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteResourceServer(data.Identifier.ValueString())
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to delete resource server",
			"Can't delete resource server "+data.Identifier.String()+": "+err.Error(),
		)
		response.Diagnostics.Append(diags...)

		return
	}

	tflog.Trace(ctx, "Deleting resource server", map[string]interface{}{
		"id": data.Id.String(),
	})

	response.State.RemoveResource(ctx)
}

// ImportState imports an existing resource into state.
// It will try to import the resource server from the old delegated cognito if not found.
func (r ResourceServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importedResourceServer central_cognito.ResourceServer

	err := r.client.ReadResourceServer(req.ID, &importedResourceServer)
	if err != nil {
		err = r.client.ImportResourceServer(req.ID, &importedResourceServer)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to import resource server",
				fmt.Sprintf(
					"The resource server could not be found in the old or new delegated cognito.\n"+
						"Underlying error: %s",
					err,
				),
			)
			return
		}
	}

	var resourceServerData ResourceServerResourceModel
	domainToState(importedResourceServer, &resourceServerData)

	resp.State.Set(ctx, &resourceServerData)
}
