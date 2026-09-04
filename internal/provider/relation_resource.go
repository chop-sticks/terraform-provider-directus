// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &relationResource{}
	_ resource.ResourceWithConfigure   = &relationResource{}
	_ resource.ResourceWithImportState = &relationResource{}
)

func RelationResource() resource.Resource { return &relationResource{} }

type relationResource struct {
	client *directus.Client
}

type relationResourceModel struct {
	Collection        types.String         `tfsdk:"collection"`
	Field             types.String         `tfsdk:"field"`
	RelatedCollection types.String         `tfsdk:"related_collection"`
	Meta              *relationMetaModel   `tfsdk:"meta"`
	Schema            *relationSchemaModel `tfsdk:"schema"`
}

type relationMetaModel struct {
	JunctionField         types.String `tfsdk:"junction_field"`
	ManyCollection        types.String `tfsdk:"many_collection"`
	ManyField             types.String `tfsdk:"many_field"`
	OneAllowedCollections types.List   `tfsdk:"one_allowed_collections"`
	OneCollection         types.String `tfsdk:"one_collection"`
	OneCollectionField    types.String `tfsdk:"one_collection_field"`
	OneDeselectAction     types.String `tfsdk:"one_deselect_action"`
	OneField              types.String `tfsdk:"one_field"`
	SortField             types.String `tfsdk:"sort_field"`
}

type relationSchemaModel struct {
	Column           types.String `tfsdk:"column"`
	ConstraintName   types.String `tfsdk:"constraint_name"`
	ForeignKeyColumn types.String `tfsdk:"foreign_key_column"`
	ForeignKeySchema types.String `tfsdk:"foreign_key_schema"`
	ForeignKeyTable  types.String `tfsdk:"foreign_key_table"`
	OnDelete         types.String `tfsdk:"on_delete"`
	OnUpdate         types.String `tfsdk:"on_update"`
	Table            types.String `tfsdk:"table"`
}

func (r *relationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_relation"
}

func (r *relationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus relation between collections (directus_relations + the underlying foreign key).",
		Attributes: map[string]schema.Attribute{
			"collection": schema.StringAttribute{
				MarkdownDescription: "The 'many' collection that holds the relational field. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"field": schema.StringAttribute{
				MarkdownDescription: "The relational field on the collection. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"related_collection": optionalComputedString("The 'one' collection this relation points to."),
			"meta": schema.SingleNestedAttribute{
				MarkdownDescription: "Directus relation metadata (directus_relations).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"junction_field":          optionalComputedString("Junction field for m2m relations."),
					"many_collection":         optionalComputedString("The 'many' collection name."),
					"many_field":              optionalComputedString("The 'many' side field name."),
					"one_allowed_collections": optionalComputedStringList("Allowed collections for m2a relations."),
					"one_collection":          optionalComputedString("The 'one' collection name."),
					"one_collection_field":    optionalComputedString("The 'one' collection field for m2a relations."),
					"one_deselect_action":     optionalComputedString("Action on deselect (nullify or delete)."),
					"one_field":               optionalComputedString("The 'one' side field name (reverse alias)."),
					"sort_field":              optionalComputedString("Field used to sort related items."),
				},
			},
			"schema": schema.SingleNestedAttribute{
				MarkdownDescription: "Underlying database foreign-key constraint.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"column":             optionalComputedString("Local column backing the relation."),
					"constraint_name":    optionalComputedString("Foreign-key constraint name."),
					"foreign_key_column": optionalComputedString("Referenced column."),
					"foreign_key_schema": optionalComputedString("Referenced schema."),
					"foreign_key_table":  optionalComputedString("Referenced table."),
					"on_delete":          optionalComputedString("ON DELETE behavior (e.g. SET NULL, CASCADE)."),
					"on_update":          optionalComputedString("ON UPDATE behavior."),
					"table":              optionalComputedString("Local table backing the relation."),
				},
			},
		},
	}
}

func (r *relationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan relationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := relationModelToClient(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateRelation(payload)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus relation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, relationToModel(ctx, created, plan.Meta != nil, plan.Schema != nil, &response.Diagnostics))...)
}

func (r *relationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state relationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	relation, err := r.client.GetRelation(state.Collection.ValueString(), state.Field.ValueString())
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus relation", err.Error())
		return
	}
	if relation == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, relationToModel(ctx, relation, state.Meta != nil, state.Schema != nil, &response.Diagnostics))...)
}

func (r *relationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan relationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := relationModelToClient(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchRelation(plan.Collection.ValueString(), plan.Field.ValueString(), payload, nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus relation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, relationToModel(ctx, updated, plan.Meta != nil, plan.Schema != nil, &response.Diagnostics))...)
}

func (r *relationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state relationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteRelation(state.Collection.ValueString(), state.Field.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus relation", err.Error())
	}
}

func (r *relationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

// ImportState parses a "collection/field" identifier into the two id attributes.
func (r *relationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
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

func relationModelToClient(ctx context.Context, m relationResourceModel, diags *diag.Diagnostics) *directus.Relation {
	relation := &directus.Relation{
		Collection:        m.Collection.ValueString(),
		Field:             m.Field.ValueString(),
		RelatedCollection: m.RelatedCollection.ValueString(),
	}

	if m.Meta != nil {
		meta := &directus.RelationMeta{
			JunctionField:      m.Meta.JunctionField.ValueString(),
			ManyCollection:     m.Meta.ManyCollection.ValueString(),
			ManyField:          m.Meta.ManyField.ValueString(),
			OneCollection:      m.Meta.OneCollection.ValueString(),
			OneCollectionField: m.Meta.OneCollectionField.ValueString(),
			OneDeselectAction:  m.Meta.OneDeselectAction.ValueString(),
			OneField:           m.Meta.OneField.ValueString(),
			SortField:          m.Meta.SortField.ValueString(),
		}
		if !m.Meta.OneAllowedCollections.IsNull() && !m.Meta.OneAllowedCollections.IsUnknown() {
			diags.Append(m.Meta.OneAllowedCollections.ElementsAs(ctx, &meta.OneAllowedCollections, false)...)
		}
		relation.Meta = meta
	}

	if m.Schema != nil {
		relation.Schema = &directus.RelationSchema{
			Column:           m.Schema.Column.ValueString(),
			ConstraintName:   m.Schema.ConstraintName.ValueString(),
			ForeignKeyColumn: m.Schema.ForeignKeyColumn.ValueString(),
			ForeignKeySchema: m.Schema.ForeignKeySchema.ValueString(),
			ForeignKeyTable:  m.Schema.ForeignKeyTable.ValueString(),
			OnDelete:         m.Schema.OnDelete.ValueString(),
			OnUpdate:         m.Schema.OnUpdate.ValueString(),
			Table:            m.Schema.Table.ValueString(),
		}
	}

	return relation
}

func relationToModel(ctx context.Context, rel *directus.Relation, includeMeta, includeSchema bool, diags *diag.Diagnostics) relationResourceModel {
	model := relationResourceModel{
		Collection:        types.StringValue(rel.Collection),
		Field:             types.StringValue(rel.Field),
		RelatedCollection: types.StringValue(rel.RelatedCollection),
	}

	if includeMeta && rel.Meta != nil {
		allowed, d := types.ListValueFrom(ctx, types.StringType, rel.Meta.OneAllowedCollections)
		diags.Append(d...)
		model.Meta = &relationMetaModel{
			JunctionField:         types.StringValue(rel.Meta.JunctionField),
			ManyCollection:        types.StringValue(rel.Meta.ManyCollection),
			ManyField:             types.StringValue(rel.Meta.ManyField),
			OneAllowedCollections: allowed,
			OneCollection:         types.StringValue(rel.Meta.OneCollection),
			OneCollectionField:    types.StringValue(rel.Meta.OneCollectionField),
			OneDeselectAction:     types.StringValue(rel.Meta.OneDeselectAction),
			OneField:              types.StringValue(rel.Meta.OneField),
			SortField:             types.StringValue(rel.Meta.SortField),
		}
	}

	if includeSchema && rel.Schema != nil {
		model.Schema = &relationSchemaModel{
			Column:           types.StringValue(rel.Schema.Column),
			ConstraintName:   types.StringValue(rel.Schema.ConstraintName),
			ForeignKeyColumn: types.StringValue(rel.Schema.ForeignKeyColumn),
			ForeignKeySchema: types.StringValue(rel.Schema.ForeignKeySchema),
			ForeignKeyTable:  types.StringValue(rel.Schema.ForeignKeyTable),
			OnDelete:         types.StringValue(rel.Schema.OnDelete),
			OnUpdate:         types.StringValue(rel.Schema.OnUpdate),
			Table:            types.StringValue(rel.Schema.Table),
		}
	}

	return model
}
