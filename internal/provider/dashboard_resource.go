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
	_ resource.Resource                = &dashboardResource{}
	_ resource.ResourceWithConfigure   = &dashboardResource{}
	_ resource.ResourceWithImportState = &dashboardResource{}
)

func DashboardResource() resource.Resource { return &dashboardResource{} }

type dashboardResource struct {
	client *directus.Client
}

type dashboardResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Icon        types.String `tfsdk:"icon"`
	Note        types.String `tfsdk:"note"`
	Color       types.String `tfsdk:"color"`
	DateCreated types.String `tfsdk:"date_created"`
	UserCreated types.String `tfsdk:"user_created"`
}

func (r *dashboardResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_dashboard"
}

func (r *dashboardResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus Insights dashboard (directus_dashboards).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned dashboard UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the dashboard.",
				Required:            true,
			},
			"icon":         optionalComputedString("Material icon name for the dashboard."),
			"note":         optionalComputedString("Description shown in the app."),
			"color":        optionalComputedString("Accent color (hex) for the dashboard."),
			"date_created": computedString("Timestamp when the dashboard was created."),
			"user_created": computedString("UUID of the user who created the dashboard."),
		},
	}
}

func (r *dashboardResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan dashboardResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateDashboard(dashboardModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus dashboard", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, dashboardToModel(created))...)
}

func (r *dashboardResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state dashboardResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	dashboard, err := r.client.GetDashboard(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus dashboard", err.Error())
		return
	}
	if dashboard == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, dashboardToModel(dashboard))...)
}

func (r *dashboardResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan dashboardResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchDashboard(plan.ID.ValueString(), dashboardModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus dashboard", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, dashboardToModel(updated))...)
}

func (r *dashboardResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state dashboardResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDashboard(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus dashboard", err.Error())
	}
}

func (r *dashboardResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *dashboardResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func dashboardModelToClient(m dashboardResourceModel) *directus.Dashboard {
	return &directus.Dashboard{
		Name:  m.Name.ValueString(),
		Icon:  m.Icon.ValueString(),
		Note:  m.Note.ValueString(),
		Color: m.Color.ValueString(),
	}
}

func dashboardToModel(d *directus.Dashboard) dashboardResourceModel {
	return dashboardResourceModel{
		ID:          types.StringValue(d.ID),
		Name:        types.StringValue(d.Name),
		Icon:        types.StringValue(d.Icon),
		Note:        types.StringValue(d.Note),
		Color:       types.StringValue(d.Color),
		DateCreated: types.StringValue(d.DateCreated),
		UserCreated: anyToStringID(d.UserCreated),
	}
}
