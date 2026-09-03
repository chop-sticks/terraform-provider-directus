// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &flowResource{}
	_ resource.ResourceWithConfigure   = &flowResource{}
	_ resource.ResourceWithImportState = &flowResource{}
)

func FlowResource() resource.Resource { return &flowResource{} }

type flowResource struct {
	client *directus.Client
}

type flowResourceModel struct {
	ID             types.String         `tfsdk:"id"`
	Name           types.String         `tfsdk:"name"`
	Icon           types.String         `tfsdk:"icon"`
	Color          types.String         `tfsdk:"color"`
	Description    types.String         `tfsdk:"description"`
	Status         types.String         `tfsdk:"status"`
	Trigger        types.String         `tfsdk:"trigger"`
	Accountability types.String         `tfsdk:"accountability"`
	Options        jsontypes.Normalized `tfsdk:"options"`
	Operation      types.String         `tfsdk:"operation"`
	DateCreated    types.String         `tfsdk:"date_created"`
	UserCreated    types.String         `tfsdk:"user_created"`
}

func (r *flowResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flow"
}

func (r *flowResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus flow (directus_flows).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned flow UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the flow.",
				Required:            true,
			},
			"icon":           optionalComputedString("Material icon name for the flow."),
			"color":          optionalComputedString("Accent color (hex) for the flow."),
			"description":    optionalComputedString("Description of the flow."),
			"status":         optionalComputedString("Flow status (active or inactive)."),
			"trigger":        optionalComputedString("Trigger type (manual, webhook, schedule, event, operation)."),
			"accountability": optionalComputedString("Accountability tracking scope (all or activity)."),
			"options":        normalizedJSONAttribute("Trigger-specific options as a JSON object."),
			"operation":      computedString("UUID of the flow's first operation."),
			"date_created":   computedString("Timestamp when the flow was created."),
			"user_created":   computedString("UUID of the user who created the flow."),
		},
	}
}

func (r *flowResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateFlow(flowModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus flow", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, flowToModel(created))...)
}

func (r *flowResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	flow, err := r.client.GetFlow(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus flow", err.Error())
		return
	}
	if flow == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, flowToModel(flow))...)
}

func (r *flowResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan flowResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchFlow(plan.ID.ValueString(), flowModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus flow", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, flowToModel(updated))...)
}

func (r *flowResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state flowResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFlow(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus flow", err.Error())
	}
}

func (r *flowResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *flowResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func flowModelToClient(m flowResourceModel, diags *diag.Diagnostics) *directus.Flow {
	return &directus.Flow{
		Name:           m.Name.ValueString(),
		Icon:           m.Icon.ValueString(),
		Color:          m.Color.ValueString(),
		Description:    m.Description.ValueString(),
		Status:         m.Status.ValueString(),
		Trigger:        m.Trigger.ValueString(),
		Accountability: m.Accountability.ValueString(),
		Options:        normalizedToMap(m.Options, diags),
	}
}

func flowToModel(f *directus.Flow) flowResourceModel {
	return flowResourceModel{
		ID:             types.StringValue(f.ID),
		Name:           types.StringValue(f.Name),
		Icon:           types.StringValue(f.Icon),
		Color:          types.StringValue(f.Color),
		Description:    types.StringValue(f.Description),
		Status:         types.StringValue(f.Status),
		Trigger:        types.StringValue(f.Trigger),
		Accountability: types.StringValue(f.Accountability),
		Options:        mapToNormalized(f.Options),
		Operation:      anyToStringID(f.Operation),
		DateCreated:    types.StringValue(f.DateCreated),
		UserCreated:    anyToStringID(f.UserCreated),
	}
}
