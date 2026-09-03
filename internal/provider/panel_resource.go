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
	_ resource.Resource                = &panelResource{}
	_ resource.ResourceWithConfigure   = &panelResource{}
	_ resource.ResourceWithImportState = &panelResource{}
)

func PanelResource() resource.Resource { return &panelResource{} }

type panelResource struct {
	client *directus.Client
}

type panelResourceModel struct {
	ID          types.String         `tfsdk:"id"`
	Dashboard   types.String         `tfsdk:"dashboard"`
	Name        types.String         `tfsdk:"name"`
	Icon        types.String         `tfsdk:"icon"`
	Color       types.String         `tfsdk:"color"`
	ShowHeader  types.Bool           `tfsdk:"show_header"`
	Note        types.String         `tfsdk:"note"`
	Type        types.String         `tfsdk:"type"`
	PositionX   types.Int64          `tfsdk:"position_x"`
	PositionY   types.Int64          `tfsdk:"position_y"`
	Width       types.Int64          `tfsdk:"width"`
	Height      types.Int64          `tfsdk:"height"`
	Options     jsontypes.Normalized `tfsdk:"options"`
	DateCreated types.String         `tfsdk:"date_created"`
	UserCreated types.String         `tfsdk:"user_created"`
}

func (r *panelResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_panel"
}

func (r *panelResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus Insights panel within a dashboard (directus_panels).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned panel UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dashboard": schema.StringAttribute{
				MarkdownDescription: "UUID of the dashboard this panel belongs to.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Panel type id (e.g. label, metric, time_series).",
				Required:            true,
			},
			"name":         optionalComputedString("Name of the panel."),
			"icon":         optionalComputedString("Material icon name for the panel."),
			"color":        optionalComputedString("Accent color (hex) for the panel."),
			"note":         optionalComputedString("Description shown in the app."),
			"show_header":  optionalComputedBool("Whether to show the panel header."),
			"position_x":   optionalComputedInt64("Panel X position on the dashboard grid."),
			"position_y":   optionalComputedInt64("Panel Y position on the dashboard grid."),
			"width":        optionalComputedInt64("Panel width in grid units."),
			"height":       optionalComputedInt64("Panel height in grid units."),
			"options":      normalizedJSONAttribute("Panel-type-specific options as a JSON object."),
			"date_created": computedString("Timestamp when the panel was created."),
			"user_created": computedString("UUID of the user who created the panel."),
		},
	}
}

func (r *panelResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan panelResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePanel(panelModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus panel", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, panelToModel(created))...)
}

func (r *panelResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state panelResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	panel, err := r.client.GetPanel(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus panel", err.Error())
		return
	}
	if panel == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, panelToModel(panel))...)
}

func (r *panelResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan panelResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchPanel(plan.ID.ValueString(), panelModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus panel", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, panelToModel(updated))...)
}

func (r *panelResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state panelResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePanel(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus panel", err.Error())
	}
}

func (r *panelResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *panelResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func panelModelToClient(m panelResourceModel, diags *diag.Diagnostics) *directus.Panel {
	return &directus.Panel{
		Dashboard:  m.Dashboard.ValueString(),
		Name:       m.Name.ValueString(),
		Icon:       m.Icon.ValueString(),
		Color:      m.Color.ValueString(),
		ShowHeader: m.ShowHeader.ValueBool(),
		Note:       m.Note.ValueString(),
		Type:       m.Type.ValueString(),
		PositionX:  int(m.PositionX.ValueInt64()),
		PositionY:  int(m.PositionY.ValueInt64()),
		Width:      int(m.Width.ValueInt64()),
		Height:     int(m.Height.ValueInt64()),
		Options:    normalizedToMap(m.Options, diags),
	}
}

func panelToModel(p *directus.Panel) panelResourceModel {
	return panelResourceModel{
		ID:          types.StringValue(p.ID),
		Dashboard:   anyToStringID(p.Dashboard),
		Name:        types.StringValue(p.Name),
		Icon:        types.StringValue(p.Icon),
		Color:       types.StringValue(p.Color),
		ShowHeader:  types.BoolValue(p.ShowHeader),
		Note:        types.StringValue(p.Note),
		Type:        types.StringValue(p.Type),
		PositionX:   types.Int64Value(int64(p.PositionX)),
		PositionY:   types.Int64Value(int64(p.PositionY)),
		Width:       types.Int64Value(int64(p.Width)),
		Height:      types.Int64Value(int64(p.Height)),
		Options:     mapToNormalized(p.Options),
		DateCreated: types.StringValue(p.DateCreated),
		UserCreated: anyToStringID(p.UserCreated),
	}
}
