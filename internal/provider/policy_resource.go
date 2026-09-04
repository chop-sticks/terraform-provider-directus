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
	_ resource.Resource                = &policyResource{}
	_ resource.ResourceWithConfigure   = &policyResource{}
	_ resource.ResourceWithImportState = &policyResource{}
)

func PolicyResource() resource.Resource { return &policyResource{} }

type policyResource struct {
	client *directus.Client
}

type policyResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Description types.String `tfsdk:"description"`
	IPAccess    types.String `tfsdk:"ip_access"`
	EnforceTFA  types.Bool   `tfsdk:"enforce_tfa"`
	AdminAccess types.Bool   `tfsdk:"admin_access"`
	AppAccess   types.Bool   `tfsdk:"app_access"`
}

func (r *policyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_policy"
}

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus access policy (directus_policies).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned policy UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the policy.",
				Required:            true,
			},
			"icon":         optionalComputedString("Material icon name for the policy."),
			"description":  optionalComputedString("Description of the policy."),
			"ip_access":    optionalComputedString("Comma-separated list of allowed IP addresses/CIDRs."),
			"enforce_tfa":  optionalComputedBool("Whether two-factor authentication is enforced for this policy."),
			"admin_access": optionalComputedBool("Whether the policy grants full administrative access."),
			"app_access":   optionalComputedBool("Whether the policy grants access to the Directus app."),
		},
	}
}

func (r *policyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan policyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePolicy(policyModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus policy", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, policyToModel(created))...)
}

func (r *policyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state policyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetPolicy(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus policy", err.Error())
		return
	}
	if policy == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, policyToModel(policy))...)
}

func (r *policyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan policyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchPolicy(plan.ID.ValueString(), policyModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus policy", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, policyToModel(updated))...)
}

func (r *policyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state policyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePolicy(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus policy", err.Error())
	}
}

func (r *policyResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *policyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func policyModelToClient(m policyResourceModel) *directus.Policy {
	return &directus.Policy{
		Name:        m.Name.ValueString(),
		Icon:        m.Icon.ValueString(),
		Description: m.Description.ValueString(),
		IPAccess:    m.IPAccess.ValueString(),
		EnforceTFA:  m.EnforceTFA.ValueBool(),
		AdminAccess: m.AdminAccess.ValueBool(),
		AppAccess:   m.AppAccess.ValueBool(),
	}
}

func policyToModel(policy *directus.Policy) policyResourceModel {
	return policyResourceModel{
		ID:          types.StringValue(policy.ID),
		Name:        types.StringValue(policy.Name),
		Icon:        types.StringValue(policy.Icon),
		Description: types.StringValue(policy.Description),
		IPAccess:    types.StringValue(policy.IPAccess),
		EnforceTFA:  types.BoolValue(policy.EnforceTFA),
		AdminAccess: types.BoolValue(policy.AdminAccess),
		AppAccess:   types.BoolValue(policy.AppAccess),
	}
}
