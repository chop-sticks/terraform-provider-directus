// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"fmt"
	"strings"

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
	_ resource.Resource                = &fieldResource{}
	_ resource.ResourceWithConfigure   = &fieldResource{}
	_ resource.ResourceWithImportState = &fieldResource{}
)

func FieldResource() resource.Resource { return &fieldResource{} }

type fieldResource struct {
	client *directus.Client
}

type fieldResourceModel struct {
	Collection types.String      `tfsdk:"collection"`
	Field      types.String      `tfsdk:"field"`
	Type       types.String      `tfsdk:"type"`
	Meta       *fieldMetaModel   `tfsdk:"meta"`
	Schema     *fieldSchemaModel `tfsdk:"schema"`
}

type fieldMetaModel struct {
	Interface         types.String         `tfsdk:"interface"`
	Display           types.String         `tfsdk:"display"`
	Note              types.String         `tfsdk:"note"`
	Width             types.String         `tfsdk:"width"`
	Group             types.String         `tfsdk:"group"`
	Hidden            types.Bool           `tfsdk:"hidden"`
	Readonly          types.Bool           `tfsdk:"readonly"`
	Required          types.Bool           `tfsdk:"required"`
	Searchable        types.Bool           `tfsdk:"searchable"`
	Sort              types.Int64          `tfsdk:"sort"`
	Special           types.List           `tfsdk:"special"`
	Options           jsontypes.Normalized `tfsdk:"options"`
	DisplayOptions    jsontypes.Normalized `tfsdk:"display_options"`
	Validation        jsontypes.Normalized `tfsdk:"validation"`
	ValidationMessage types.String         `tfsdk:"validation_message"`
}

type fieldSchemaModel struct {
	DataType     types.String         `tfsdk:"data_type"`
	DefaultValue jsontypes.Normalized `tfsdk:"default_value"`
	MaxLength    types.Int64          `tfsdk:"max_length"`
	IsNullable   types.Bool           `tfsdk:"is_nullable"`
	IsUnique     types.Bool           `tfsdk:"is_unique"`
	IsPrimaryKey types.Bool           `tfsdk:"is_primary_key"`
	Comment      types.String         `tfsdk:"comment"`
}

func (r *fieldResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_field"
}

func (r *fieldResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus field within a collection (directus_fields + the underlying column).",
		Attributes: map[string]schema.Attribute{
			"collection": schema.StringAttribute{
				MarkdownDescription: "Collection the field belongs to. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field": schema.StringAttribute{
				MarkdownDescription: "Field (column) name. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Directus field type (e.g. string, integer, uuid, boolean, json).",
				Required:            true,
			},
			"meta": schema.SingleNestedAttribute{
				MarkdownDescription: "Directus field metadata (directus_fields).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"interface":          optionalComputedString("Interface id used to edit the field."),
					"display":            optionalComputedString("Display id used to render the field."),
					"note":               optionalComputedString("Help note shown under the field."),
					"width":              optionalComputedString("Field width in the app form (half, full, ...)."),
					"group":              optionalComputedString("Field group this field is nested under."),
					"hidden":             optionalComputedBool("Whether the field is hidden in the app."),
					"readonly":           optionalComputedBool("Whether the field is read-only in the app."),
					"required":           optionalComputedBool("Whether the field is required in the app."),
					"searchable":         optionalComputedBool("Whether the field is included in global search."),
					"sort":               optionalComputedInt64("Sort order of the field in the app."),
					"special":            optionalComputedStringList("Special Directus behaviors (e.g. uuid, cast-json, m2o)."),
					"options":            normalizedJSONAttribute("Interface-specific options as a JSON object."),
					"display_options":    normalizedJSONAttribute("Display-specific options as a JSON object."),
					"validation":         normalizedJSONAttribute("Validation filter rules as a JSON object."),
					"validation_message": optionalComputedString("Custom validation error message."),
				},
			},
			"schema": schema.SingleNestedAttribute{
				MarkdownDescription: "Underlying database column definition.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"data_type":      computedString("Database column data type (server-derived from type)."),
					"default_value":  normalizedJSONAttribute("Column default value as JSON (string, number, bool, or null)."),
					"max_length":     optionalComputedInt64("Maximum length for string columns."),
					"is_nullable":    optionalComputedBool("Whether the column allows NULL."),
					"is_unique":      optionalComputedBool("Whether the column has a unique constraint."),
					"is_primary_key": computedBool("Whether the column is the primary key."),
					"comment":        optionalComputedString("Database-level column comment."),
				},
			},
		},
	}
}

func (r *fieldResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan fieldResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	field := fieldModelToClient(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateField(plan.Collection.ValueString(), field, nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus field", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, fieldToModel(ctx, created, plan.Meta != nil, plan.Schema != nil, &response.Diagnostics))...)
}

func (r *fieldResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state fieldResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	field, err := r.client.GetFieldByCollectionAndName(state.Collection.ValueString(), state.Field.ValueString())
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus field", err.Error())
		return
	}
	if field == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, fieldToModel(ctx, field, state.Meta != nil, state.Schema != nil, &response.Diagnostics))...)
}

func (r *fieldResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan fieldResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	field := fieldModelToClient(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchField(plan.Collection.ValueString(), plan.Field.ValueString(), field, nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus field", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, fieldToModel(ctx, updated, plan.Meta != nil, plan.Schema != nil, &response.Diagnostics))...)
}

func (r *fieldResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state fieldResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteField(state.Collection.ValueString(), state.Field.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus field", err.Error())
	}
}

func (r *fieldResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

// ImportState parses a "collection/field" identifier into the two id attributes.
func (r *fieldResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	collection, field, ok := strings.Cut(request.ID, "/")
	if !ok || collection == "" || field == "" {
		response.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import id in the form \"collection/field\", got %q.", request.ID),
		)
		return
	}
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("collection"), collection)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("field"), field)...)
}

// --- model <-> client mapping ---

func fieldModelToClient(ctx context.Context, m fieldResourceModel, diags *diag.Diagnostics) *directus.Field {
	field := &directus.Field{
		Collection: m.Collection.ValueString(),
		Field:      m.Field.ValueString(),
		Type:       m.Type.ValueString(),
	}

	if m.Meta != nil {
		meta := &directus.FieldMeta{
			Interface:         m.Meta.Interface.ValueString(),
			Display:           m.Meta.Display.ValueString(),
			Note:              m.Meta.Note.ValueString(),
			Width:             m.Meta.Width.ValueString(),
			Group:             m.Meta.Group.ValueString(),
			Hidden:            m.Meta.Hidden.ValueBool(),
			Readonly:          m.Meta.Readonly.ValueBool(),
			Required:          m.Meta.Required.ValueBool(),
			Searchable:        m.Meta.Searchable.ValueBool(),
			Sort:              int(m.Meta.Sort.ValueInt64()),
			Options:           normalizedToMap(m.Meta.Options, diags),
			DisplayOptions:    normalizedToMap(m.Meta.DisplayOptions, diags),
			Validation:        normalizedToMap(m.Meta.Validation, diags),
			ValidationMessage: m.Meta.ValidationMessage.ValueString(),
		}
		if !m.Meta.Special.IsNull() && !m.Meta.Special.IsUnknown() {
			diags.Append(m.Meta.Special.ElementsAs(ctx, &meta.Special, false)...)
		}
		field.Meta = meta
	}

	if m.Schema != nil {
		field.Schema = &directus.FieldSchema{
			DataType:     m.Schema.DataType.ValueString(),
			DefaultValue: normalizedToAny(m.Schema.DefaultValue, diags),
			MaxLength:    int(m.Schema.MaxLength.ValueInt64()),
			IsNullable:   m.Schema.IsNullable.ValueBool(),
			IsUnique:     m.Schema.IsUnique.ValueBool(),
			Comment:      m.Schema.Comment.ValueString(),
		}
	}

	return field
}

func fieldToModel(ctx context.Context, f *directus.Field, includeMeta, includeSchema bool, diags *diag.Diagnostics) fieldResourceModel {
	model := fieldResourceModel{
		Collection: types.StringValue(f.Collection),
		Field:      types.StringValue(f.Field),
		Type:       types.StringValue(f.Type),
	}

	if includeMeta && f.Meta != nil {
		special, d := types.ListValueFrom(ctx, types.StringType, f.Meta.Special)
		diags.Append(d...)
		model.Meta = &fieldMetaModel{
			Interface:         types.StringValue(f.Meta.Interface),
			Display:           types.StringValue(f.Meta.Display),
			Note:              types.StringValue(f.Meta.Note),
			Width:             types.StringValue(f.Meta.Width),
			Group:             types.StringValue(f.Meta.Group),
			Hidden:            types.BoolValue(f.Meta.Hidden),
			Readonly:          types.BoolValue(f.Meta.Readonly),
			Required:          types.BoolValue(f.Meta.Required),
			Searchable:        types.BoolValue(f.Meta.Searchable),
			Sort:              types.Int64Value(int64(f.Meta.Sort)),
			Special:           special,
			Options:           mapToNormalized(f.Meta.Options),
			DisplayOptions:    mapToNormalized(f.Meta.DisplayOptions),
			Validation:        mapToNormalized(f.Meta.Validation),
			ValidationMessage: types.StringValue(f.Meta.ValidationMessage),
		}
	}

	if includeSchema && f.Schema != nil {
		model.Schema = &fieldSchemaModel{
			DataType:     types.StringValue(f.Schema.DataType),
			DefaultValue: anyToNormalized(f.Schema.DefaultValue),
			MaxLength:    types.Int64Value(int64(f.Schema.MaxLength)),
			IsNullable:   types.BoolValue(f.Schema.IsNullable),
			IsUnique:     types.BoolValue(f.Schema.IsUnique),
			IsPrimaryKey: types.BoolValue(f.Schema.IsPrimaryKey),
			Comment:      types.StringValue(f.Schema.Comment),
		}
	}

	return model
}
