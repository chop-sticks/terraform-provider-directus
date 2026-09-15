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
	_ resource.Resource                = &roleResource{}
	_ resource.ResourceWithConfigure   = &roleResource{}
	_ resource.ResourceWithImportState = &roleResource{}
)

func RoleResource() resource.Resource { return &roleResource{} }

type roleResource struct {
	client *directus.Client
}

type roleResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Description types.String `tfsdk:"description"`
	Parent      types.String `tfsdk:"parent"`
}

func (r *roleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_role"
}

func (r *roleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus role (directus_roles).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned role UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the role.",
				Required:            true,
			},
			"icon":        optionalComputedString("Material icon name for the role."),
			"description": optionalComputedString("Description of the role."),
			"parent":      optionalComputedString("UUID of the parent role, for nesting."),
		},
	}
}

func (r *roleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan roleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateRole(roleModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus role", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, roleToModel(created))...)
}

func (r *roleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state roleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	role, err := r.client.GetRole(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus role", err.Error())
		return
	}
	if role == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, roleToModel(role))...)
}

func (r *roleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan roleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchRole(plan.ID.ValueString(), roleModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus role", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, roleToModel(updated))...)
}

func (r *roleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state roleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRole(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus role", err.Error())
	}
}

func (r *roleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *roleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func roleModelToClient(m roleResourceModel) *directus.Role {
	role := &directus.Role{
		Name:        m.Name.ValueString(),
		Icon:        m.Icon.ValueString(),
		Description: m.Description.ValueString(),
	}
	if !m.Parent.IsNull() && !m.Parent.IsUnknown() {
		role.Parent = m.Parent.ValueString()
	}
	return role
}

func roleToModel(role *directus.Role) roleResourceModel {
	return roleResourceModel{
		ID:          types.StringValue(role.ID),
		Name:        types.StringValue(role.Name),
		Icon:        types.StringValue(role.Icon),
		Description: types.StringValue(role.Description),
		Parent:      anyToStringID(role.Parent),
	}
}
