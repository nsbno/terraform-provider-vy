package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

func scopeName(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "/" + name
}

func NewResourceServerScopeResource() resource.Resource {
	return &ResourceServerScopeResource{}
}

type ResourceServerScopeResource struct {
	client *central_cognito.Client
}

type ResourceServerScopeResourceModel struct {
	Id             types.String `tfsdk:"id"`
	ResourceServer types.String `tfsdk:"resource_server"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
}

func (r ResourceServerScopeResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_resource_server_scope"
}

func (r ResourceServerScopeResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "A single scope on a resource server, managed independently of the " +
			"`vy_resource_server` resource itself. Look up the resource server with the " +
			"`vy_resource_server` data source.",

		Attributes: map[string]schema.Attribute{
			// id is required by the SDKv2 testing framework.
			// See https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resource_server": schema.StringAttribute{
				MarkdownDescription: "The identifier of the resource server this scope belongs to. " +
					"Typically the `identifier` of a `vy_resource_server` resource, or the " +
					"`identifier` of a `vy_resource_server` data source.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"namespace": schema.StringAttribute{
				MarkdownDescription: "An optional namespace for this scope, e.g. the name of the " +
					"microservice/domain that owns it. Helps avoid name collisions when several " +
					"teams share one resource server. The scope name sent to Cognito is " +
					"`namespace/name`, or just `name` if this is omitted.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for this scope",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "A description of what this scope is for",
				Required:            true,
			},
		},
	}
}

func (r *ResourceServerScopeResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	configuration, ok := request.ProviderData.(*VyProviderConfiguration)

	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *VyProviderConfiguration, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = configuration.CognitoClient
}

func (r ResourceServerScopeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data ResourceServerScopeResourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerId := data.ResourceServer.ValueString()
	name := scopeName(data.Namespace.ValueString(), data.Name.ValueString())

	_, err := r.client.CreateResourceServerScope(resourceServerId, central_cognito.Scope{
		Name:        name,
		Description: data.Description.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to create scope",
			fmt.Sprintf("Could not add scope %q to resource server %q: %s", name, resourceServerId, err.Error()),
		)
		return
	}

	data.Id = types.StringValue(resourceServerId + "/" + name)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerScopeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data ResourceServerScopeResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerId := data.ResourceServer.ValueString()
	name := scopeName(data.Namespace.ValueString(), data.Name.ValueString())

	// There's no endpoint to read a single scope, so we read the whole resource server
	// and look for it in the scope list.
	var server central_cognito.ResourceServer
	if err := r.client.ReadResourceServer(resourceServerId, &server); err != nil {
		response.Diagnostics.AddError(
			"Unable to read resource server",
			fmt.Sprintf("Could not read resource server %q for scope %q: %s", resourceServerId, name, err.Error()),
		)
		return
	}

	found := false
	for _, existing := range server.Scopes {
		if existing.Name == name {
			data.Description = types.StringValue(existing.Description)
			found = true
			break
		}
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	data.Id = types.StringValue(resourceServerId + "/" + name)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerScopeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data ResourceServerScopeResourceModel

	diags := request.Plan.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerId := data.ResourceServer.ValueString()
	name := scopeName(data.Namespace.ValueString(), data.Name.ValueString())

	_, err := r.client.UpdateResourceServerScope(resourceServerId, central_cognito.Scope{
		Name:        name,
		Description: data.Description.ValueString(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			"Unable to update scope",
			fmt.Sprintf("Could not update scope %q on resource server %q: %s", name, resourceServerId, err.Error()),
		)
		return
	}

	data.Id = types.StringValue(resourceServerId + "/" + name)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (r ResourceServerScopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data ResourceServerScopeResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerId := data.ResourceServer.ValueString()
	name := scopeName(data.Namespace.ValueString(), data.Name.ValueString())

	if _, err := r.client.DeleteResourceServerScope(resourceServerId, name); err != nil {
		response.Diagnostics.AddError(
			"Unable to delete scope",
			fmt.Sprintf("Could not remove scope %q from resource server %q: %s", name, resourceServerId, err.Error()),
		)
		return
	}

	response.State.RemoveResource(ctx)
}

// ImportState imports an existing scope into state.
// Use the format "<resource_server_identifier>,<scope_name>", where <scope_name> is the full
// scope name as stored remotely (i.e. including any "namespace/" prefix, if present). The
// imported resource will have `name` set to the full scope name and `namespace` left empty;
// split them manually in config afterwards if desired.
func (r ResourceServerScopeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected import identifier",
			fmt.Sprintf("Expected import identifier with format: <resource_server_identifier>,<scope_name>. Got: %q", req.ID),
		)
		return
	}

	resourceServerId := parts[0]
	name := parts[1]

	var server central_cognito.ResourceServer
	if err := r.client.ReadResourceServer(resourceServerId, &server); err != nil {
		resp.Diagnostics.AddError(
			"Unable to import scope",
			fmt.Sprintf("Could not read resource server %q: %s", resourceServerId, err.Error()),
		)
		return
	}

	for _, existing := range server.Scopes {
		if existing.Name == name {
			data := ResourceServerScopeResourceModel{
				Id:             types.StringValue(resourceServerId + "/" + name),
				ResourceServer: types.StringValue(resourceServerId),
				Namespace:      types.StringNull(),
				Name:           types.StringValue(name),
				Description:    types.StringValue(existing.Description),
			}
			resp.State.Set(ctx, &data)
			return
		}
	}

	resp.Diagnostics.AddError(
		"Unable to import scope",
		fmt.Sprintf("Resource server %q has no scope named %q", resourceServerId, name),
	)
}
