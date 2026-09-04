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
	_ resource.Resource                = &operationResource{}
	_ resource.ResourceWithConfigure   = &operationResource{}
	_ resource.ResourceWithImportState = &operationResource{}
)

func OperationResource() resource.Resource { return &operationResource{} }

type operationResource struct {
	client *directus.Client
}

type operationResourceModel struct {
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Key         types.String         `tfsdk:"key"`
	Type        types.String         `tfsdk:"type"`
	PositionX   types.Int64          `tfsdk:"position_x"`
	PositionY   types.Int64          `tfsdk:"position_y"`
	Options     jsontypes.Normalized `tfsdk:"options"`
	Resolve     types.String         `tfsdk:"resolve"`
	Reject      types.String         `tfsdk:"reject"`
	Flow        types.String         `tfsdk:"flow"`
	DateCreated types.String         `tfsdk:"date_created"`
	UserCreated types.String         `tfsdk:"user_created"`
}

func (r *operationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_operation"
}

func (r *operationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus flow operation (directus_operations).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned operation UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"flow": schema.StringAttribute{
				MarkdownDescription: "UUID of the flow this operation belongs to.",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Unique key of the operation within the flow.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Operation type id (e.g. log, mail, request, condition).",
				Required:            true,
			},
			"name":         optionalComputedString("Name of the operation."),
			"position_x":   optionalComputedInt64("Operation X position on the flow grid."),
			"position_y":   optionalComputedInt64("Operation Y position on the flow grid."),
			"options":      normalizedJSONAttribute("Operation-type-specific options as a JSON object."),
			"resolve":      optionalComputedString("UUID of the operation to run on success."),
			"reject":       optionalComputedString("UUID of the operation to run on failure."),
			"date_created": computedString("Timestamp when the operation was created."),
			"user_created": computedString("UUID of the user who created the operation."),
		},
	}
}

func (r *operationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan operationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := operationModelToClient(plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateOperation(payload, nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus operation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, operationToModel(created))...)
}

func (r *operationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state operationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	operation, err := r.client.GetOperation(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus operation", err.Error())
		return
	}
	if operation == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, operationToModel(operation))...)
}

func (r *operationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan operationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := operationModelToClient(plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchOperation(plan.ID.ValueString(), payload, nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus operation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, operationToModel(updated))...)
}

func (r *operationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state operationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteOperation(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus operation", err.Error())
	}
}

func (r *operationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *operationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func operationModelToClient(m operationResourceModel, diags *diag.Diagnostics) *directus.Operation {
	operation := &directus.Operation{
		Name:      m.Name.ValueString(),
		Key:       m.Key.ValueString(),
		Type:      m.Type.ValueString(),
		PositionX: int(m.PositionX.ValueInt64()),
		PositionY: int(m.PositionY.ValueInt64()),
		Options:   normalizedToMap(m.Options, diags),
		Flow:      m.Flow.ValueString(),
	}
	if !m.Resolve.IsNull() && !m.Resolve.IsUnknown() {
		operation.Resolve = m.Resolve.ValueString()
	}
	if !m.Reject.IsNull() && !m.Reject.IsUnknown() {
		operation.Reject = m.Reject.ValueString()
	}
	return operation
}

func operationToModel(o *directus.Operation) operationResourceModel {
	return operationResourceModel{
		ID:          types.StringValue(o.ID),
		Name:        types.StringValue(o.Name),
		Key:         types.StringValue(o.Key),
		Type:        types.StringValue(o.Type),
		PositionX:   types.Int64Value(int64(o.PositionX)),
		PositionY:   types.Int64Value(int64(o.PositionY)),
		Options:     mapToNormalized(o.Options),
		Resolve:     anyToStringID(o.Resolve),
		Reject:      anyToStringID(o.Reject),
		Flow:        anyToStringID(o.Flow),
		DateCreated: types.StringValue(o.DateCreated),
		UserCreated: anyToStringID(o.UserCreated),
	}
}
