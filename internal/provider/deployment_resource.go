package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/nsbno/terraform-provider-vy/internal/enroll_account"
)

type deploymentResourceType struct{}

func (t deploymentResourceType) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Register the current AWS account into the deployment service",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type: types.StringType,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.UseStateForUnknown(),
				},
				Computed: true,
			},
			"topics": {
				MarkdownDescription: "All the topics that are required for",
				Required:            true,
				Attributes: tfsdk.SingleNestedAttributes(
					map[string]tfsdk.Attribute{
						"trigger_events": {
							MarkdownDescription: "SNS topic ARN for all pipeline start triggers",
							Required:            true,
							Type:                types.StringType,
						},
						"pipeline_events": {
							MarkdownDescription: "SNS topic ARN for all events from the pipeline",
							Required:            true,
							Type:                types.StringType,
						},
					},
				),
				PlanModifiers: tfsdk.AttributePlanModifiers{
					tfsdk.RequiresReplace(),
				},
			},
		},
	}, nil
}

func (t deploymentResourceType) NewResource(ctx context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return deploymentResource{
		provider: provider,
	}, diags
}

type topics struct {
	TriggerEvents  types.String `tfsdk:"trigger_events"`
	PipelineEvents types.String `tfsdk:"pipeline_events"`
}

type deploymentResourceData struct {
	Id     types.String `tfsdk:"id"`
	Topics topics       `tfsdk:"topics"`
}

type deploymentResource struct {
	provider provider
}

func deployAccountDomainToState(account *enroll_account.Account, data *deploymentResourceData) {
	data.Id.Value = account.AccountId
	data.Topics.TriggerEvents.Value = account.Topics.TriggerEvents.Arn
	data.Topics.PipelineEvents.Value = account.Topics.PipelineEvents.Arn
}

func (d deploymentResource) Create(ctx context.Context, request tfsdk.CreateResourceRequest, response *tfsdk.CreateResourceResponse) {
	var data deploymentResourceData

	diags := request.Config.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	created, err := d.provider.EnrollAccountClient.CreateAccount(
		enroll_account.Topics{
			TriggerEvents:  enroll_account.Topic{Arn: data.Topics.TriggerEvents.Value},
			PipelineEvents: enroll_account.Topic{Arn: data.Topics.PipelineEvents.Value},
		},
	)
	if err != nil {
		response.Diagnostics.AddError(
			"Could not enroll account for deployments",
			fmt.Sprintf("%s", err.Error()),
		)

		return
	}

	var createdData deploymentResourceData
	deployAccountDomainToState(created, &createdData)

	diags = response.State.Set(ctx, &createdData)
	response.Diagnostics.Append(diags...)
}

func (d deploymentResource) Read(ctx context.Context, request tfsdk.ReadResourceRequest, response *tfsdk.ReadResourceResponse) {
	var data deploymentResourceData

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	var readData enroll_account.Account
	err := d.provider.EnrollAccountClient.ReadAccount(&readData)

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

func (d deploymentResource) Update(ctx context.Context, request tfsdk.UpdateResourceRequest, response *tfsdk.UpdateResourceResponse) {
	var data deploymentResourceData

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

func (d deploymentResource) Delete(ctx context.Context, request tfsdk.DeleteResourceRequest, response *tfsdk.DeleteResourceResponse) {
	var data deploymentResourceData

	diags := request.State.Get(ctx, &data)
	response.Diagnostics.Append(diags...)

	if response.Diagnostics.HasError() {
		return
	}

	err := d.provider.EnrollAccountClient.DeleteAccount()
	if err != nil {
		response.Diagnostics.AddError(
			"Could not delete account",
			err.Error(),
		)

		return
	}

	response.State.RemoveResource(ctx)
}
