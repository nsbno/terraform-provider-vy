package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/nsbno/terraform-provider-vy/internal/central_cognito"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ validator.String = frontendOrBackendValidator{}

type frontendOrBackendValidator struct{}

func (t frontendOrBackendValidator) Description(ctx context.Context) string {
	return "type must be either 'frontend' or 'backend'"
}

func (t frontendOrBackendValidator) MarkdownDescription(ctx context.Context) string {
	return "type must be either `frontend` or `backend`"
}

func (t frontendOrBackendValidator) ValidateString(ctx context.Context, request validator.StringRequest, response *validator.StringResponse) {
	var str = request.ConfigValue

	if str.IsUnknown() || str.IsNull() {
		return
	}

	if str.ValueString() != "frontend" && str.ValueString() != "backend" {
		response.Diagnostics.AddAttributeError(
			request.Path,
			"Invalid app client type",
			fmt.Sprintf("The app client must either be 'frontend' or 'backend'. Got: '%s'.", str.ValueString()),
		)

		return
	}
}

func NewAppClientResource() resource.Resource {
	return &AppClientResource{}
}

type AppClientResource struct {
	client *central_cognito.Client
}

type AppClientResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Scopes         []string     `tfsdk:"scopes"`
	Type           types.String `tfsdk:"type"`
	CallbackUrls   []string     `tfsdk:"callback_urls"`
	LogoutUrls     []string     `tfsdk:"logout_urls"`
	GenerateSecret types.Bool   `tfsdk:"generate_secret"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
}

func (r *AppClientResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_app_client"
}

func (r *AppClientResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "App clients are the user pool authentication resources attached to your app. " +
			"Use an app client to configure the permitted authentication actions towards a resource server.",

		Attributes: map[string]schema.Attribute{
			// id is required by the SDKv2 testing framework.
			// See https://www.terraform.io/plugin/framework/acctests#implement-id-attribute
			// TODO: Later, this is probably going to be the generated ID from cognito.
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of this app client",
				Required:            true,
			},
			"scopes": schema.SetAttribute{
				MarkdownDescription: "Scopes that this client has access to",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The use-case for this app client. Used to automatically add OAuth options. " +
					"Must be either `frontend` or `backend`.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					frontendOrBackendValidator{},
				},
			},
			"callback_urls": schema.ListAttribute{
				MarkdownDescription: "Callback URLs to use. Used together with `type` set to `frontend`.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"logout_urls": schema.ListAttribute{
				MarkdownDescription: "Logout URLs to use. Used together with `type` set to `frontend`.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"generate_secret": schema.BoolAttribute{
				MarkdownDescription: "Should a secret be generated? Automatically set by `type`, but you're able to override it with this option.",
				Optional:            true,
				Computed:            true, // The backend can change it if it is not set by the user.
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The ID used for your client to authenticate itself. ",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "A secret used for your client to authenticate itself. " +
					"Only populated when using the `backend` type.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (c *AppClientResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (ac AppClientResourceModel) toDomain(domain *central_cognito.AppClient) {
	domain.Name = ac.Name.ValueString()
	domain.Scopes = ac.Scopes
	domain.Type = ac.Type.ValueString()

	// The remote expects it to always be a list.
	if ac.CallbackUrls == nil {
		domain.CallbackUrls = []string{}
	} else {
		domain.CallbackUrls = ac.CallbackUrls
	}

	if ac.LogoutUrls == nil {
		domain.LogoutUrls = []string{}
	} else {
		domain.LogoutUrls = ac.LogoutUrls
	}

	if !ac.GenerateSecret.IsNull() {
		value := ac.GenerateSecret.ValueBool()
		domain.GenerateSecret = &value
	}
}

func appClientResourceDataFromDomain(domain central_cognito.AppClient, state *AppClientResourceModel) {
	state.Id = types.StringValue(domain.Name)
	state.Name = types.StringValue(domain.Name)
	state.Scopes = domain.Scopes
	state.Type = types.StringValue(domain.Type)

	// If the config is empty on our side, terraform expects a null, not an empty list.
	// There is probably a better way to handle this, but I can't find anything in the docs.
	if len(domain.CallbackUrls) == 0 {
		state.CallbackUrls = nil
	} else {
		state.CallbackUrls = domain.CallbackUrls
	}

	if len(domain.LogoutUrls) == 0 {
		state.LogoutUrls = nil
	} else {
		state.LogoutUrls = domain.LogoutUrls
	}

	if domain.GenerateSecret != nil {
		state.GenerateSecret = types.BoolValue(*domain.GenerateSecret)
	}

	if domain.ClientId != nil {
		state.ClientId = types.StringValue(*domain.ClientId)
	}

	if domain.ClientSecret != nil {
		state.ClientSecret = types.StringValue(*domain.ClientSecret)
	}
}

func (r AppClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppClientResourceModel

	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = data.Name

	var appClient central_cognito.AppClient
	data.toDomain(&appClient)

	var createdAppClient, err = r.client.CreateAppClient(appClient)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Could not create app client",
			fmt.Sprintf("App client with name %s could not be created: %s", appClient.Name, err.Error()),
		)
		resp.Diagnostics.Append(diags...)

		return
	}

	var createdAppClientResource AppClientResourceModel
	appClientResourceDataFromDomain(*createdAppClient, &createdAppClientResource)

	diags = resp.State.Set(ctx, &createdAppClientResource)
	resp.Diagnostics.Append(diags...)
}

func (r AppClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppClientResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	var server central_cognito.AppClient
	err := r.client.ReadAppClient(data.Name.ValueString(), &server)
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to read app client",
			fmt.Sprintf("Can't read app client %s from remote: %s ", data.Name.ValueString(), err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	var newState AppClientResourceModel
	appClientResourceDataFromDomain(server, &newState)

	diags = resp.State.Set(ctx, &newState)
	resp.Diagnostics.Append(diags...)
}

func (r AppClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppClientResourceModel
	var state AppClientResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	// I don't know why, but after the upgrade to the new terraform provider framework,
	// null values are still marked as unkown in the plan.
	if data.ClientSecret.IsUnknown() && state.ClientSecret.IsNull() {
		data.ClientSecret = state.ClientSecret
	}

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = data.Name

	var appClient central_cognito.AppClient
	data.toDomain(&appClient)

	err := r.client.UpdateAppClient(central_cognito.AppClientUpdateRequest{
		Name:         appClient.Name,
		Scopes:       appClient.Scopes,
		CallbackUrls: appClient.CallbackUrls,
		LogoutUrls:   appClient.LogoutUrls,
	})
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to update app client",
			fmt.Sprintf("Can't update app client %s in remote: %s ", data.Name.ValueString(), err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r AppClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppClientResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAppClient(data.Name.ValueString())
	if err != nil {
		diags = diag.Diagnostics{}
		diags.AddError(
			"Unable to delete app client",
			fmt.Sprintf("Can't delete app client %s in remote: %s ", data.Name.ValueString(), err.Error()),
		)
		resp.Diagnostics.Append(diags...)
		return
	}

	resp.State.RemoveResource(ctx)
}

// ImportState imports an existing app client into the state.
// If it doesn't find a app client in the system, it will try to import from old delegated cognito.
//
// Because the new system uses names as its primary key, it is dual function.
// To import an existing app client, use the name of the app client.
// To import from the old system, the `client_id` must be used.
func (r AppClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	var importedAppClient central_cognito.AppClient

	err := r.client.ReadAppClient(req.ID, &importedAppClient)

	if err != nil {
		err = r.client.ImportAppClient(req.ID, &importedAppClient)

		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to import app client",
				fmt.Sprintf("The app client could not be found in the new or old delegated Cognito.\nUnderlying error: %s", err),
			)
			return
		}
	}

	var appClientData AppClientResourceModel
	appClientResourceDataFromDomain(importedAppClient, &appClientData)

	resp.State.Set(ctx, &appClientData)
}
