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

// Interface assertions: collectionResource implements the full resource surface
// (CRUD + Configure + ImportState). Every resource in this provider follows the
// same shape — use this file as the reference template.
var (
	_ resource.Resource                = &collectionResource{}
	_ resource.ResourceWithConfigure   = &collectionResource{}
	_ resource.ResourceWithImportState = &collectionResource{}
)

func CollectionResource() resource.Resource { return &collectionResource{} }

type collectionResource struct {
	client *directus.Client
}

type collectionResourceModel struct {
	Collection types.String           `tfsdk:"collection"`
	Meta       *collectionMetaModel   `tfsdk:"meta"`
	Schema     *collectionSchemaModel `tfsdk:"schema"`
}

// collectionMetaModel exposes the commonly-managed directus_collections meta
// scalars as typed attributes (the hybrid modeling decision). Fields that
// Directus rarely surfaces for IaC are intentionally omitted; extend here as
// needed.
type collectionMetaModel struct {
	Icon            types.String `tfsdk:"icon"`
	Note            types.String `tfsdk:"note"`
	Color           types.String `tfsdk:"color"`
	DisplayTemplate types.String `tfsdk:"display_template"`
	Hidden          types.Bool   `tfsdk:"hidden"`
	Singleton       types.Bool   `tfsdk:"singleton"`
	SortField       types.String `tfsdk:"sort_field"`
	Group           types.String `tfsdk:"group"`
	Collapse        types.String `tfsdk:"collapse"`
}

// collectionSchemaModel mirrors the underlying database table info. Name is
// server-managed (it equals the collection name); schema and comment are
// optional.
type collectionSchemaModel struct {
	Name    types.String `tfsdk:"name"`
	Schema  types.String `tfsdk:"schema"`
	Comment types.String `tfsdk:"comment"`
}

func (r *collectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_collection"
}

func (r *collectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus collection (its metadata and underlying database table).",
		Attributes: map[string]schema.Attribute{
			"collection": schema.StringAttribute{
				MarkdownDescription: "Name of the collection (also the database table name). Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"meta": schema.SingleNestedAttribute{
				MarkdownDescription: "Directus-managed presentation metadata (directus_collections).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"icon":             optionalComputedString("Material icon name shown in the app."),
					"note":             optionalComputedString("Description shown in the app."),
					"color":            optionalComputedString("Accent color (hex) shown in the app."),
					"display_template": optionalComputedString("Display template for item previews."),
					"hidden":           optionalComputedBool("Whether the collection is hidden in the app."),
					"singleton":        optionalComputedBool("Whether the collection is a singleton."),
					"sort_field":       optionalComputedString("Field used for manual sorting."),
					"group":            optionalComputedString("Parent collection for nesting in the app."),
					"collapse":         optionalComputedString("Default collapse behavior in the app (open, closed, locked)."),
				},
			},
			"schema": schema.SingleNestedAttribute{
				MarkdownDescription: "Underlying database table info.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name":    computedString("Table name (equals the collection name)."),
					"schema":  optionalComputedString("Database schema the table belongs to."),
					"comment": optionalComputedString("Database-level comment on the table."),
				},
			},
		},
	}
}

func (r *collectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan collectionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Directus creates a table-less "folder" collection (and rejects it on some
	// versions) when no schema is provided. Always send a schema object so a
	// real table is created; an unmanaged schema is simply empty.
	req := &directus.CollectionRequest{
		Collection: plan.Collection.ValueString(),
		Meta:       metaModelToRequest(plan.Meta),
		Schema:     schemaModelToRequest(plan.Schema),
	}
	if req.Schema == nil {
		req.Schema = &directus.CollectionSchema{}
	}

	created, err := r.client.CreateCollection(req, nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus collection", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(created, plan.Meta != nil, plan.Schema != nil))...)
}

func (r *collectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state collectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	col, err := r.client.GetCollectionByName(state.Collection.ValueString())
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus collection", err.Error())
		return
	}
	if col == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(col, state.Meta != nil, state.Schema != nil))...)
}

func (r *collectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan collectionResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	req := &directus.CollectionRequest{
		Collection: plan.Collection.ValueString(),
		Meta:       metaModelToRequest(plan.Meta),
		Schema:     schemaModelToRequest(plan.Schema),
	}

	updated, err := r.client.PatchCollection(plan.Collection.ValueString(), req, nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus collection", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(updated, plan.Meta != nil, plan.Schema != nil))...)
}

func (r *collectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state collectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCollection(state.Collection.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus collection", err.Error())
	}
}

func (r *collectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *collectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("collection"), request, response)
}

// --- model <-> client mapping ---

func metaModelToRequest(m *collectionMetaModel) *directus.CollectionMeta {
	if m == nil {
		return nil
	}
	return &directus.CollectionMeta{
		Icon:            m.Icon.ValueString(),
		Note:            m.Note.ValueString(),
		Color:           m.Color.ValueString(),
		DisplayTemplate: m.DisplayTemplate.ValueString(),
		Hidden:          m.Hidden.ValueBool(),
		Singleton:       m.Singleton.ValueBool(),
		SortField:       m.SortField.ValueString(),
		Group:           m.Group.ValueString(),
		Collapse:        m.Collapse.ValueString(),
	}
}

func schemaModelToRequest(s *collectionSchemaModel) *directus.CollectionSchema {
	if s == nil {
		return nil
	}
	return &directus.CollectionSchema{
		Name:    s.Name.ValueString(),
		Schema:  s.Schema.ValueString(),
		Comment: s.Comment.ValueString(),
	}
}

// collectionToModel maps the client type into the resource model. includeMeta
// and includeSchema control whether the nested blocks are populated: they are
// only tracked in state when the user manages them, so an unmanaged (omitted)
// block stays null and does not produce a perpetual diff.
func collectionToModel(col *directus.Collection, includeMeta, includeSchema bool) collectionResourceModel {
	model := collectionResourceModel{
		Collection: types.StringValue(col.Collection),
	}
	if includeMeta && col.Meta != nil {
		model.Meta = &collectionMetaModel{
			Icon:            types.StringValue(col.Meta.Icon),
			Note:            types.StringValue(col.Meta.Note),
			Color:           types.StringValue(col.Meta.Color),
			DisplayTemplate: types.StringValue(col.Meta.DisplayTemplate),
			Hidden:          types.BoolValue(col.Meta.Hidden),
			Singleton:       types.BoolValue(col.Meta.Singleton),
			SortField:       types.StringValue(col.Meta.SortField),
			Group:           types.StringValue(col.Meta.Group),
			Collapse:        types.StringValue(col.Meta.Collapse),
		}
	}
	if includeSchema && col.Schema != nil {
		model.Schema = &collectionSchemaModel{
			Name:    types.StringValue(col.Schema.Name),
			Schema:  types.StringValue(col.Schema.Schema),
			Comment: types.StringValue(col.Schema.Comment),
		}
	}
	return model
}
