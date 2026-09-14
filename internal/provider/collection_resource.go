// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"fmt"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	Fields     []collectionFieldModel `tfsdk:"fields"`
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
	PreviewURL      types.String `tfsdk:"preview_url"`
}

// collectionSchemaModel mirrors the underlying database table info. Name is
// server-managed (it equals the collection name); schema and comment are
// optional.
type collectionSchemaModel struct {
	Name    types.String `tfsdk:"name"`
	Schema  types.String `tfsdk:"schema"`
	Comment types.String `tfsdk:"comment"`
}

// collectionFieldModel describes an initial field created alongside the
// collection (POST /collections "fields"). It is a create-only input: Directus
// never returns "fields" on collection reads, so these values are stored as
// configured and preserved across Read/Update rather than reconciled from the
// server. Changing them forces a new resource.
type collectionFieldModel struct {
	Field  types.String                `tfsdk:"field"`
	Type   types.String                `tfsdk:"type"`
	Meta   *collectionFieldMetaModel   `tfsdk:"meta"`
	Schema *collectionFieldSchemaModel `tfsdk:"schema"`
}

type collectionFieldMetaModel struct {
	Interface         types.String         `tfsdk:"interface"`
	Display           types.String         `tfsdk:"display"`
	Note              types.String         `tfsdk:"note"`
	Width             types.String         `tfsdk:"width"`
	Group             types.String         `tfsdk:"group"`
	Hidden            types.Bool           `tfsdk:"hidden"`
	Readonly          types.Bool           `tfsdk:"readonly"`
	Required          types.Bool           `tfsdk:"required"`
	Sort              types.Int64          `tfsdk:"sort"`
	Special           types.List           `tfsdk:"special"`
	Options           jsontypes.Normalized `tfsdk:"options"`
	DisplayOptions    jsontypes.Normalized `tfsdk:"display_options"`
	Validation        jsontypes.Normalized `tfsdk:"validation"`
	ValidationMessage types.String         `tfsdk:"validation_message"`
}

type collectionFieldSchemaModel struct {
	DataType         types.String         `tfsdk:"data_type"`
	DefaultValue     jsontypes.Normalized `tfsdk:"default_value"`
	MaxLength        types.Int64          `tfsdk:"max_length"`
	NumericPrecision types.Int64          `tfsdk:"numeric_precision"`
	NumericScale     types.Int64          `tfsdk:"numeric_scale"`
	IsNullable       types.Bool           `tfsdk:"is_nullable"`
	IsUnique         types.Bool           `tfsdk:"is_unique"`
	IsPrimaryKey     types.Bool           `tfsdk:"is_primary_key"`
	HasAutoIncrement types.Bool           `tfsdk:"has_auto_increment"`
	Comment          types.String         `tfsdk:"comment"`
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
					"collapse":         optionalComputedStringDefault("Default collapse behavior in the app (open, closed, locked).", "open"),
					"preview_url":      optionalComputedString("URL template used to preview items (e.g. a live site URL)."),
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
			"fields": schema.ListNestedAttribute{
				MarkdownDescription: "Initial fields to create alongside the collection (POST /collections). " +
					"Typically used to define the collection's primary key/ID field. Only applied at " +
					"creation and never read back from Directus, so changing this forces a new resource.",
				Optional: true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"field": schema.StringAttribute{
							MarkdownDescription: "Field (column) name, e.g. \"id\".",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Directus field type (e.g. integer, uuid, string).",
							Required:            true,
						},
						"meta": schema.SingleNestedAttribute{
							MarkdownDescription: "Directus field metadata (directus_fields).",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"interface":          optionalString("Interface id used to edit the field."),
								"display":            optionalString("Display id used to render the field."),
								"note":               optionalString("Help note shown under the field."),
								"width":              optionalString("Field width in the app form (half, full, ...)."),
								"group":              optionalString("Field group this field is nested under."),
								"hidden":             optionalBool("Whether the field is hidden in the app."),
								"readonly":           optionalBool("Whether the field is read-only in the app."),
								"required":           optionalBool("Whether the field is required in the app."),
								"sort":               optionalInt64("Sort order of the field in the app."),
								"special":            optionalStringList("Special Directus behaviors (e.g. uuid, cast-json, m2o)."),
								"options":            optionalNormalizedJSON("Interface-specific options as a JSON object."),
								"display_options":    optionalNormalizedJSON("Display-specific options as a JSON object."),
								"validation":         optionalNormalizedJSON("Validation filter rules as a JSON object."),
								"validation_message": optionalString("Custom validation error message."),
							},
						},
						"schema": schema.SingleNestedAttribute{
							MarkdownDescription: "Underlying database column definition.",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"data_type":          optionalString("Database column data type (e.g. integer, uuid, varchar)."),
								"default_value":      optionalNormalizedJSON("Column default value as JSON (string, number, bool, or null)."),
								"max_length":         optionalInt64("Maximum length for string columns."),
								"numeric_precision":  optionalInt64("Numeric precision for numeric columns."),
								"numeric_scale":      optionalInt64("Numeric scale for numeric columns."),
								"is_nullable":        optionalBool("Whether the column allows NULL."),
								"is_unique":          optionalBool("Whether the column has a unique constraint."),
								"is_primary_key":     optionalBool("Whether the column is the primary key."),
								"has_auto_increment": optionalBool("Whether the column auto-increments (auto-increment integer ID)."),
								"comment":            optionalString("Database-level column comment."),
							},
						},
					},
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
		Fields:     collectionFieldsToRequest(ctx, plan.Fields, plan.Collection.ValueString(), &response.Diagnostics),
	}
	if response.Diagnostics.HasError() {
		return
	}
	if req.Schema == nil {
		req.Schema = &directus.CollectionSchema{}
	}

	// Serialize with all other schema writes: creating a collection with inline
	// fields races destructively against concurrent DDL (see schemaMu).
	schemaMu.Lock()
	defer schemaMu.Unlock()

	created, err := createCollectionVerified(r.client, req)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus collection", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(created, plan.Meta != nil, plan.Schema != nil, plan.Fields))...)
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

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(col, state.Meta != nil, state.Schema != nil, state.Fields))...)
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

	schemaMu.Lock()
	updated, err := r.client.PatchCollection(plan.Collection.ValueString(), req, nil)
	schemaMu.Unlock()
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus collection", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(updated, plan.Meta != nil, plan.Schema != nil, plan.Fields))...)
}

func (r *collectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state collectionResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	schemaMu.Lock()
	err := r.client.DeleteCollection(state.Collection.ValueString())
	schemaMu.Unlock()
	if err != nil {
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
		PreviewURL:      m.PreviewURL.ValueString(),
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
// block stays null and does not produce a perpetual diff. fields is carried
// through verbatim: Directus never returns a collection's create-time "fields",
// so the configured/prior-state value is preserved as-is.
func collectionToModel(col *directus.Collection, includeMeta, includeSchema bool, fields []collectionFieldModel) collectionResourceModel {
	model := collectionResourceModel{
		Collection: types.StringValue(col.Collection),
		Fields:     fields,
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
			PreviewURL:      types.StringValue(col.Meta.PreviewURL),
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

// collectionFieldsToRequest builds the create-only "fields" payload for POST
// /collections. Each entry's collection and schema table name are stamped from
// the parent collection so the payload is internally consistent, matching what
// Directus expects when creating fields inline.
func collectionFieldsToRequest(ctx context.Context, fields []collectionFieldModel, collection string, diags *diag.Diagnostics) []directus.Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]directus.Field, 0, len(fields))
	for _, f := range fields {
		field := directus.Field{
			Collection: collection,
			Field:      f.Field.ValueString(),
			Type:       f.Type.ValueString(),
		}
		if f.Meta != nil {
			meta := &directus.FieldMeta{
				Collection:        collection,
				Field:             f.Field.ValueString(),
				Interface:         f.Meta.Interface.ValueString(),
				Display:           f.Meta.Display.ValueString(),
				Note:              f.Meta.Note.ValueString(),
				Width:             f.Meta.Width.ValueString(),
				Group:             f.Meta.Group.ValueString(),
				Hidden:            f.Meta.Hidden.ValueBool(),
				Readonly:          f.Meta.Readonly.ValueBool(),
				Required:          f.Meta.Required.ValueBool(),
				Sort:              int(f.Meta.Sort.ValueInt64()),
				Options:           normalizedToMap(f.Meta.Options, diags),
				DisplayOptions:    normalizedToMap(f.Meta.DisplayOptions, diags),
				Validation:        normalizedToMap(f.Meta.Validation, diags),
				ValidationMessage: f.Meta.ValidationMessage.ValueString(),
			}
			if !f.Meta.Special.IsNull() && !f.Meta.Special.IsUnknown() {
				diags.Append(f.Meta.Special.ElementsAs(ctx, &meta.Special, false)...)
			}
			field.Meta = meta
		}
		if f.Schema != nil {
			field.Schema = &directus.FieldSchema{
				Name:             f.Field.ValueString(),
				Table:            collection,
				DataType:         f.Schema.DataType.ValueString(),
				DefaultValue:     normalizedToAny(f.Schema.DefaultValue, diags),
				MaxLength:        int(f.Schema.MaxLength.ValueInt64()),
				NumericPrecision: int(f.Schema.NumericPrecision.ValueInt64()),
				NumericScale:     int(f.Schema.NumericScale.ValueInt64()),
				IsNullable:       f.Schema.IsNullable.ValueBool(),
				IsUnique:         f.Schema.IsUnique.ValueBool(),
				IsPrimaryKey:     f.Schema.IsPrimaryKey.ValueBool(),
				HasAutoIncrement: f.Schema.HasAutoIncrement.ValueBool(),
				Comment:          f.Schema.Comment.ValueString(),
			}
		}
		out = append(out, field)
	}
	return out
}

// createCollectionVerified creates a collection and, when the request carries
// inline fields, confirms they were actually created. Directus can return 200
// from POST /collections while silently dropping the inline fields, leaving a
// primary-key-less table it then reports as "does not exist" (403) — a state
// that cannot be repaired in place (the missing primary key cannot be added
// afterward). The caller holds schemaMu, so no concurrent DDL interferes and a
// dropped-field result is a transient server fault that delete+recreate clears.
// The broken collection is always removed before returning an error, leaving no
// orphan to block a later apply.
func createCollectionVerified(client *directus.Client, req *directus.CollectionRequest) (*directus.Collection, error) {
	const attempts = 3
	var lastErr error
	for range attempts {
		created, err := client.CreateCollection(req, nil)
		if err != nil {
			return nil, err
		}
		if len(req.Fields) == 0 {
			return created, nil
		}
		if lastErr = collectionMissingFields(client, req.Collection, req.Fields); lastErr == nil {
			return created, nil
		}
		_ = client.DeleteCollection(req.Collection)
	}
	return nil, fmt.Errorf("collection %q was created without its configured fields after %d attempts (Directus dropped the inline field payload): %w", req.Collection, attempts, lastErr)
}

// collectionMissingFields returns a non-nil error when any requested field is
// absent from the collection. A collection lacking its primary key answers
// field reads with a 403 "does not exist", which surfaces here as that error.
func collectionMissingFields(client *directus.Client, collection string, want []directus.Field) error {
	got, err := client.GetFieldsByCollection(collection)
	if err != nil {
		return err
	}
	present := make(map[string]struct{}, len(got))
	for _, f := range got {
		present[f.Field] = struct{}{}
	}
	for _, f := range want {
		if _, ok := present[f.Field]; !ok {
			return fmt.Errorf("field %q was not created", f.Field)
		}
	}
	return nil
}
