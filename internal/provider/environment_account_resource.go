package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/enroll_account"
)

func NewEnvironmentAccountResource() resource.Resource {
	return &EnvironmentAccountResource{}
}

type EnvironmentAccountResource struct {
	client *enroll_account.Client
}

type EnvironmentAccountResourceModel struct {
	Id             types.String `tfsdk:"id"`
	OwnerAccountId types.String `tfsdk:"owner_account_id"`
}

func (e EnvironmentAccountResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_environment_account"
}

func (e EnvironmentAccountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Register the current AWS account as an environment for the deployment service",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"owner_account_id": schema.StringAttribute{
				MarkdownDescription: "The deployment account that owns this account. Aka the service account.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (e EnvironmentAccountResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	e.client = configuration.EnrollAccountClient
}

func environmentAccountDomainToState(domain *enroll_account.EnvironmentAccount, e *EnvironmentAccountResourceModel) {
	e.Id = types.StringValue(domain.AccountId)
	e.OwnerAccountId = types.StringValue(domain.OwnerAccountId)
}

func (e EnvironmentAccountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data EnvironmentAccountResourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	registered, err := e.client.RegisterEnvironmentAccount(data.OwnerAccountId.ValueString())
	if err != nil {
		response.Diagnostics.AddError(
			"Could not enroll environment account",
			fmt.Sprintf("%s", err.Error()),
		)
	}

	var registeredData EnvironmentAccountResourceModel
	environmentAccountDomainToState(registered, &registeredData)

	diags = response.State.Set(ctx, &registeredData)
	response.Diagnostics.Append(diags...)
}

func (e EnvironmentAccountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data EnvironmentAccountResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	var readData enroll_account.EnvironmentAccount
	err := e.client.ReadEnvironmentAccount(&readData)

	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read environment account information",
			err.Error(),
		)
		return
	}

	environmentAccountDomainToState(&readData, &data)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (e EnvironmentAccountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data EnvironmentAccountResourceModel

	diags := request.Plan.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	// NOTE: There are no fields that should allow the client to update.
	//		 The only action is create, read or delete.

	diags = response.State.Set(ctx, data)
	response.Diagnostics.Append(diags...)
}

func (e EnvironmentAccountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data EnvironmentAccountResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := e.client.DeleteEnvironmentAccount()
	if err != nil {
		response.Diagnostics.AddError(
			"Could not delete account",
			err.Error(),
		)

		return
	}

	response.State.RemoveResource(ctx)
}
