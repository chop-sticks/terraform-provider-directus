// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &accessResource{}
	_ resource.ResourceWithConfigure   = &accessResource{}
	_ resource.ResourceWithImportState = &accessResource{}
)

func AccessResource() resource.Resource { return &accessResource{} }

type accessResource struct {
	client *directus.Client
}

// accessResourceModel maps directus_access: the junction assigning a policy to
// a user or a role. Exactly one of user/role is set; the other stays null.
type accessResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Policy types.String `tfsdk:"policy"`
	Role   types.String `tfsdk:"role"`
	User   types.String `tfsdk:"user"`
	Sort   types.Int64  `tfsdk:"sort"`
}

func (r *accessResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_access"
}

func (r *accessResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus access record (directus_access): assigns an access policy to a user or a role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned access record UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "UUID of the policy being assigned.",
				Required:            true,
			},
			"role": optionalComputedString("UUID of the role the policy is assigned to. Set either role or user, not both."),
			"user": optionalComputedString("UUID of the user the policy is assigned to. Set either user or role, not both."),
			"sort": optionalComputedInt64("Manual sort order of the assignment."),
		},
	}
}

func (r *accessResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan accessResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateAccess(accessModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus access", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, accessToModel(created))...)
}

func (r *accessResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state accessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	access, err := r.client.GetAccess(state.ID.ValueString(), nil)
	if err != nil {
		if accessGone(r.client, state.ID.ValueString(), err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus access", err.Error())
		return
	}
	if access == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, accessToModel(access))...)
}

func (r *accessResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan accessResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchAccess(plan.ID.ValueString(), accessModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus access", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, accessToModel(updated))...)
}

func (r *accessResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state accessResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAccess(state.ID.ValueString()); err != nil {
		if accessGone(r.client, state.ID.ValueString(), err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus access", err.Error())
	}
}

func (r *accessResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *accessResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func accessModelToClient(m accessResourceModel) *directus.Access {
	access := &directus.Access{
		Policy: m.Policy.ValueString(),
	}
	if !m.Role.IsNull() && !m.Role.IsUnknown() {
		access.Role = m.Role.ValueString()
	}
	if !m.User.IsNull() && !m.User.IsUnknown() {
		access.User = m.User.ValueString()
	}
	if !m.Sort.IsNull() && !m.Sort.IsUnknown() {
		s := int(m.Sort.ValueInt64())
		access.Sort = &s
	}
	return access
}

func accessToModel(a *directus.Access) accessResourceModel {
	m := accessResourceModel{
		ID:     types.StringValue(a.ID),
		Policy: anyToStringID(a.Policy),
		Role:   anyToStringID(a.Role),
		User:   anyToStringID(a.User),
		Sort:   types.Int64Null(),
	}
	if a.Sort != nil {
		m.Sort = types.Int64Value(int64(*a.Sort))
	}
	return m
}
