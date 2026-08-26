package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"
)

// cognitoScopeNameMaxLength is Cognito's max length for the combined
// scopeName field ("namespace.name")
const cognitoScopeNameMaxLength = 256

// resourceServerScopeLocks serializes scope changes per resource server
var resourceServerScopeLocks sync.Map // map[string]*sync.Mutex

func lockResourceServerScopes(ctx context.Context, identifier string) func() {
	value, _ := resourceServerScopeLocks.LoadOrStore(identifier, &sync.Mutex{})
	mu := value.(*sync.Mutex)

	tflog.Debug(ctx, "Acquiring resource server scope lock", map[string]interface{}{"resource_server": identifier})
	mu.Lock()
	tflog.Debug(ctx, "Acquired resource server scope lock", map[string]interface{}{"resource_server": identifier})

	return func() {
		mu.Unlock()
		tflog.Debug(ctx, "Released resource server scope lock", map[string]interface{}{"resource_server": identifier})
	}
}

// cognitoScopeNamePartPattern matches a valid namespace or name segment,
// per Cognito's scopeName constraint (excludes "/", '"', '\' and space).
var cognitoScopeNamePartPattern = regexp.MustCompile(`^[\x21\x23-\x2E\x30-\x5B\x5D-\x7E]+$`)

// scopeName builds the scope name sent to Cognito, joined with "." since
// Cognito's scopeName field rejects "/".
func scopeName(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "." + name
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
			"`vy_resource_server` data source. Do not use this at the same time as a non-empty " +
			"`scopes` list on a [`vy_resource_server`](resource_server.md) for the same resource " +
			"server.",

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
				MarkdownDescription: "`namespace` is optional, e.g. your microservice or domain " +
					"name. It is combined with `name` as `namespace.name`, allowing teams sharing " +
					"this resource server to avoid name collisions. A `/` separator isn't used " +
					"here because Cognito's scope name field rejects that character.",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						cognitoScopeNamePartPattern,
						`must match Cognito's allowed scope name characters (letters, digits, and most punctuation, but not "/", '"', '\' or space)`,
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "A name for this scope",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						cognitoScopeNamePartPattern,
						`must match Cognito's allowed scope name characters (letters, digits, and most punctuation, but not "/", '"', '\' or space)`,
					),
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

// ValidateConfig checks the combined "namespace.name" against Cognito's length limit
func (r ResourceServerScopeResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data ResourceServerScopeResourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() || data.Namespace.IsUnknown() || data.Name.IsUnknown() {
		return
	}

	name := scopeName(data.Namespace.ValueString(), data.Name.ValueString())
	if len(name) > cognitoScopeNameMaxLength {
		response.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Scope name too long",
			fmt.Sprintf(
				"The combined scope name %q (from \"namespace.name\") is %d characters, "+
					"but Cognito allows at most %d.",
				name, len(name), cognitoScopeNameMaxLength,
			),
		)
	}
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

	unlock := lockResourceServerScopes(ctx, resourceServerId)
	defer unlock()

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

	unlock := lockResourceServerScopes(ctx, resourceServerId)
	defer unlock()

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

	unlock := lockResourceServerScopes(ctx, resourceServerId)
	defer unlock()

	if _, err := r.client.DeleteResourceServerScope(resourceServerId, name); err != nil {
		response.Diagnostics.AddError(
			"Unable to delete scope",
			fmt.Sprintf("Could not remove scope %q from resource server %q: %s", name, resourceServerId, err.Error()),
		)
		return
	}

	response.State.RemoveResource(ctx)
}

// ImportState imports an existing scope, given "<resource_server_identifier>,<scope_name>".
// The full scope_name (including any "namespace." prefix) is imported into `name`, with
// `namespace` left empty
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
