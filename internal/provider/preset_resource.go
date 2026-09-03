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
	_ resource.Resource                = &presetResource{}
	_ resource.ResourceWithConfigure   = &presetResource{}
	_ resource.ResourceWithImportState = &presetResource{}
)

func PresetResource() resource.Resource { return &presetResource{} }

type presetResource struct {
	client *directus.Client
}

type presetResourceModel struct {
	ID              types.Int64          `tfsdk:"id"`
	Bookmark        types.String         `tfsdk:"bookmark"`
	User            types.String         `tfsdk:"user"`
	Role            types.String         `tfsdk:"role"`
	Collection      types.String         `tfsdk:"collection"`
	Search          types.String         `tfsdk:"search"`
	Layout          types.String         `tfsdk:"layout"`
	LayoutQuery     jsontypes.Normalized `tfsdk:"layout_query"`
	LayoutOptions   jsontypes.Normalized `tfsdk:"layout_options"`
	RefreshInterval types.Int64          `tfsdk:"refresh_interval"`
	Filter          jsontypes.Normalized `tfsdk:"filter"`
	Icon            types.String         `tfsdk:"icon"`
	Color           types.String         `tfsdk:"color"`
}

func (r *presetResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_preset"
}

func (r *presetResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus preset — a saved bookmark or default collection view (directus_presets).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Server-assigned preset id.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"bookmark":         optionalComputedString("Bookmark name. Empty for a collection default view."),
			"user":             optionalComputedString("UUID of the user this preset applies to (omit for global)."),
			"role":             optionalComputedString("UUID of the role this preset applies to (omit for global)."),
			"collection":       optionalComputedString("Collection the preset targets."),
			"search":           optionalComputedString("Search query stored in the preset."),
			"layout":           optionalComputedString("Layout id (e.g. tabular, cards)."),
			"layout_query":     normalizedJSONAttribute("Per-layout query state as a JSON object."),
			"layout_options":   normalizedJSONAttribute("Per-layout display options as a JSON object."),
			"refresh_interval": optionalComputedInt64("Auto-refresh interval in seconds."),
			"filter":           normalizedJSONAttribute("Filter rules as a JSON object."),
			"icon":             optionalComputedString("Material icon name for the bookmark."),
			"color":            optionalComputedString("Accent color (hex) for the bookmark."),
		},
	}
}

func (r *presetResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan presetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreatePreset(presetModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus preset", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, presetToModel(created))...)
}

func (r *presetResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state presetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	preset, err := r.client.GetPreset(int(state.ID.ValueInt64()), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus preset", err.Error())
		return
	}
	if preset == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, presetToModel(preset))...)
}

func (r *presetResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan presetResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchPreset(int(plan.ID.ValueInt64()), presetModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus preset", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, presetToModel(updated))...)
}

func (r *presetResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state presetResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePreset(int(state.ID.ValueInt64())); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus preset", err.Error())
	}
}

func (r *presetResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *presetResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importInt64ID(ctx, request, response)
}

// --- model <-> client mapping ---

func presetModelToClient(m presetResourceModel, diags *diag.Diagnostics) *directus.Preset {
	preset := &directus.Preset{
		Bookmark:        m.Bookmark.ValueString(),
		Collection:      m.Collection.ValueString(),
		Search:          m.Search.ValueString(),
		Layout:          m.Layout.ValueString(),
		RefreshInterval: int(m.RefreshInterval.ValueInt64()),
		Icon:            m.Icon.ValueString(),
		Color:           m.Color.ValueString(),
		LayoutQuery:     normalizedToMap(m.LayoutQuery, diags),
		LayoutOptions:   normalizedToMap(m.LayoutOptions, diags),
		Filter:          normalizedToMap(m.Filter, diags),
	}
	if !m.User.IsNull() && !m.User.IsUnknown() {
		preset.User = m.User.ValueString()
	}
	if !m.Role.IsNull() && !m.Role.IsUnknown() {
		preset.Role = m.Role.ValueString()
	}
	return preset
}

func presetToModel(p *directus.Preset) presetResourceModel {
	return presetResourceModel{
		ID:              types.Int64Value(int64(p.ID)),
		Bookmark:        types.StringValue(p.Bookmark),
		User:            anyToStringID(p.User),
		Role:            anyToStringID(p.Role),
		Collection:      types.StringValue(p.Collection),
		Search:          types.StringValue(p.Search),
		Layout:          types.StringValue(p.Layout),
		LayoutQuery:     mapToNormalized(p.LayoutQuery),
		LayoutOptions:   mapToNormalized(p.LayoutOptions),
		RefreshInterval: types.Int64Value(int64(p.RefreshInterval)),
		Filter:          mapToNormalized(p.Filter),
		Icon:            types.StringValue(p.Icon),
		Color:           types.StringValue(p.Color),
	}
}
