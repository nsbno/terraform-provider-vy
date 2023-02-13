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

func NewDeploymentResource() resource.Resource {
	return &DeploymentResource{}
}

type DeploymentResource struct {
	client *enroll_account.Client
}

type DeploymentResourceModel struct {
	Id           types.String `tfsdk:"id"`
	SlackChannel types.String `tfsdk:"slack_channel"`
}

func (d DeploymentResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_deployment_account"
}

func (d DeploymentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Register the current AWS account into the deployment service",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"slack_channel": schema.StringAttribute{
				MarkdownDescription: "A Slack channel where info about deployments go",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (d *DeploymentResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	d.client = configuration.EnrollAccountClient
}

func deployAccountDomainToState(account *enroll_account.Account, data *DeploymentResourceModel) {
	data.Id = types.StringValue(account.AccountId)
	data.SlackChannel = types.StringValue(account.SlackChannel)
}

func (d DeploymentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data DeploymentResourceModel

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	created, err := d.client.CreateAccount(
		data.SlackChannel.ValueString(),
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Could not enroll account for deployments",
			fmt.Sprintf("%s", err.Error()),
		)

		return
	}

	var createdData DeploymentResourceModel
	deployAccountDomainToState(created, &createdData)

	diags = response.State.Set(ctx, &createdData)
	response.Diagnostics.Append(diags...)
}

func (d DeploymentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data DeploymentResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var readData enroll_account.Account
	err := d.client.ReadAccount(&readData)

	if err != nil {
		response.Diagnostics.AddError(
			"Unable to read deployment account information",
			err.Error(),
		)
		return
	}

	deployAccountDomainToState(&readData, &data)

	diags = response.State.Set(ctx, &data)
	response.Diagnostics.Append(diags...)
}

func (d DeploymentResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data DeploymentResourceModel

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

func (d DeploymentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data DeploymentResourceModel

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := d.client.DeleteAccount()
	if err != nil {
		response.Diagnostics.AddError(
			"Could not delete account",
			err.Error(),
		)

		return
	}

	response.State.RemoveResource(ctx)
}
