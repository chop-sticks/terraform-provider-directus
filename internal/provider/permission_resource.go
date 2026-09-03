// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &permissionResource{}
	_ resource.ResourceWithConfigure   = &permissionResource{}
	_ resource.ResourceWithImportState = &permissionResource{}
)

func PermissionResource() resource.Resource { return &permissionResource{} }

type permissionResource struct {
	client *directus.Client
}

type permissionResourceModel struct {
	ID          types.Int64          `tfsdk:"id"`
	Policy      types.String         `tfsdk:"policy"`
	Collection  types.String         `tfsdk:"collection"`
	Action      types.String         `tfsdk:"action"`
	Permissions jsontypes.Normalized `tfsdk:"permissions"`
	Validation  jsontypes.Normalized `tfsdk:"validation"`
	Presets     jsontypes.Normalized `tfsdk:"presets"`
	Fields      types.List           `tfsdk:"fields"`
}

func (r *permissionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_permission"
}

func (r *permissionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus permission rule attached to a policy (directus_permissions). Custom (non-full-access) rules require a licensed instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Server-assigned permission id.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"policy": optionalComputedString("UUID of the policy this permission belongs to (omit for public)."),
			"collection": schema.StringAttribute{
				MarkdownDescription: "Collection the permission applies to.",
				Required:            true,
			},
			"action": schema.StringAttribute{
				MarkdownDescription: "Action allowed (create, read, update, delete, share).",
				Required:            true,
			},
			"permissions": normalizedJSONAttribute("Filter rules restricting which items the action applies to (JSON)."),
			"validation":  normalizedJSONAttribute("Validation rules for create/update payloads (JSON)."),
			"presets":     normalizedJSONAttribute("Preset values applied on create (JSON)."),
			"fields":      optionalComputedStringList("Allowed fields for the action ([\"*\"] for all)."),
		},
	}
}

func (r *permissionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan permissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePermission(permissionModelToClient(ctx, plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus permission", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, permissionToModel(ctx, created, &response.Diagnostics))...)
}

func (r *permissionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state permissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	permission, err := r.client.GetPermission(int(state.ID.ValueInt64()), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus permission", err.Error())
		return
	}
	if permission == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, permissionToModel(ctx, permission, &response.Diagnostics))...)
}

func (r *permissionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan permissionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchPermission(int(plan.ID.ValueInt64()), permissionModelToClient(ctx, plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus permission", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, permissionToModel(ctx, updated, &response.Diagnostics))...)
}

func (r *permissionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state permissionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePermission(int(state.ID.ValueInt64())); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus permission", err.Error())
	}
}

func (r *permissionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *permissionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importInt64ID(ctx, request, response)
}

// --- model <-> client mapping ---

func permissionModelToClient(ctx context.Context, m permissionResourceModel, diags *diag.Diagnostics) *directus.Permission {
	permission := &directus.Permission{
		Collection:  m.Collection.ValueString(),
		Action:      m.Action.ValueString(),
		Permissions: normalizedToMap(m.Permissions, diags),
		Validation:  normalizedToMap(m.Validation, diags),
		Presets:     normalizedToMap(m.Presets, diags),
	}
	if !m.Policy.IsNull() && !m.Policy.IsUnknown() {
		permission.Policy = m.Policy.ValueString()
	}
	if !m.Fields.IsNull() && !m.Fields.IsUnknown() {
		diags.Append(m.Fields.ElementsAs(ctx, &permission.Fields, false)...)
	}
	return permission
}

func permissionToModel(ctx context.Context, p *directus.Permission, diags *diag.Diagnostics) permissionResourceModel {
	fields, d := types.ListValueFrom(ctx, types.StringType, p.Fields)
	diags.Append(d...)
	return permissionResourceModel{
		ID:          types.Int64Value(int64(p.ID)),
		Policy:      anyToStringID(p.Policy),
		Collection:  types.StringValue(p.Collection),
		Action:      types.StringValue(p.Action),
		Permissions: mapToNormalized(p.Permissions),
		Validation:  mapToNormalized(p.Validation),
		Presets:     mapToNormalized(p.Presets),
		Fields:      fields,
	}
}
