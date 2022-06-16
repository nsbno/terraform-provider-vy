package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

type resourceServerType struct{}

func (t resourceServerType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "A cognito resource server",

		Attributes: map[string]tfsdk.Attribute{
			// id is required by the SDKv2 testing framework.
			// See https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			"id": {
				Type: types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Computed: true,
			},
			"identifier": {
				MarkdownDescription: "The identity of this resource server",
				Required:            true,
				Type:                types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
			"name": {
				MarkdownDescription: "The name of this resource server",
				Required:            true,
				Type:                types.StringType,
			},
			"scopes": {
				MarkdownDescription: "Scopes for this resource server",
				Optional:            true,
				Attributes: tfsdk.SetNestedAttributes(
					map[string]tfsdk.Attribute{
						"name": {
							MarkdownDescription: "A name for this scope",
							Required:            true,
							Type:                types.StringType,
						},
						"description": {
							MarkdownDescription: "A description of what this scope is for",
							Required:            true,
							Type:                types.StringType,
						},
					},
					tfsdk.SetNestedAttributesOptions{},
				),
				PlanModifiers: []tfsdk.AttributePlanModifier{
					tfsdk.RequiresReplace(),
				},
			},
		},
	}, nil
}

func (t resourceServerType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return resourceServer{
		provider: provider,
	}, diags
}

type resourceServerData struct {
	Id         types.String `tfsdk:"id"`
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Scopes     []scope      `tfsdk:"scopes"`
}

type scope struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type resourceServer struct {
	provider provider
}

func stateToDomain(state resourceServerData, domain *central_cognito.ResourceServer) {
	domain.Identifier = state.Identifier.Value
	domain.Name = state.Name.Value

	domain.Scopes = []central_cognito.Scope{}

	for _, state_scope := range state.Scopes {
		domain.Scopes = append(domain.Scopes, central_cognito.Scope{
			Name:        state_scope.Name.Value,
			Description: state_scope.Description.Value,
		})
	}
}

func domainToState(domain central_cognito.ResourceServer, state *resourceServerData) {
	// See schema's comment about id to see why we do this.
	state.Id.Value = domain.Identifier
	state.Id.Null = false
	state.Identifier.Value = domain.Identifier
	state.Name.Value = domain.Name

	state.Scopes = []scope{}
	for _, domain_scope := range domain.Scopes {
		state_scope := scope{}

		state_scope.Name.Value = domain_scope.Name
		state_scope.Description.Value = domain_scope.Description

		state.Scopes = append(state.Scopes, state_scope)
	}

	if len(state.Scopes) == 0 {
		// Terraform thinks this has changed if we keep the scopes as an empty list.
		state.Scopes = nil
	}
}

func (r resourceServer) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	var data resourceServerData

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Id.Value = data.Identifier.Value
	data.Id.Null = false

	var server central_cognito.ResourceServer
	stateToDomain(data, &server)

	err := r.provider.Client.CreateResourceServer(server)
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

func (r resourceServer) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
	var data resourceServerData

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var server central_cognito.ResourceServer
	err := r.provider.Client.ReadResourceServer(data.Identifier.Value, &server)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to read resource server",
			"Can't read resource server "+data.Identifier.Value+" from remote: "+err.Error(),
		)
		response.Diagnostics.Append(diags...)
		return
	}

	domainToState(server, &data)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r resourceServer) Update(ctx context.Context, request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
	var data resourceServerData

	diags := request.Plan.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Id.Value = data.Identifier.Value
	data.Id.Null = false

	var server central_cognito.ResourceServer
	stateToDomain(data, &server)

	err := r.provider.Client.UpdateResourceServer(central_cognito.ResourceServerUpdateRequest{
		Identifier: server.Identifier,
		Name:       server.Name,
		Scopes:     server.Scopes,
	})
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to update resource server",
			"Can't update resource server "+data.Identifier.Value+": "+err.Error(),
		)
		response.Diagnostics.Append(diags...)

		return
	}

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r resourceServer) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	var data resourceServerData

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := r.provider.Client.DeleteResourceServer(data.Identifier.Value)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to delete resource server",
			"Can't delete resource server "+data.Identifier.Value+": "+err.Error(),
		)
		response.Diagnostics.Append(diags...)

		return
	}

	tflog.Trace(ctx, "Deleting resource server", map[string]interface{}{
		"id": data.Id.Value,
	})

	response.State.RemoveResource(ctx)
}

func (r resourceServer) ImportState(ctx context.Context, request tfsdk.ImportResourceStateRequest, response *tfsdk.ImportResourceStateResponse) {
	//TODO implement me
	panic("implement me")
}
