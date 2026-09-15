// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// collectionDataSourceModel is the data source's own model. It intentionally
// does not reuse collectionResourceModel: the data source exposes a read-only
// subset (no create-only "fields", no preview_url) and the framework requires a
// struct that matches its schema 1:1 in both directions.
type collectionDataSourceModel struct {
	Collection types.String                     `tfsdk:"collection"`
	Meta       *collectionDataSourceMetaModel   `tfsdk:"meta"`
	Schema     *collectionDataSourceSchemaModel `tfsdk:"schema"`
}

type collectionDataSourceMetaModel struct {
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

type collectionDataSourceSchemaModel struct {
	Name    types.String `tfsdk:"name"`
	Schema  types.String `tfsdk:"schema"`
	Comment types.String `tfsdk:"comment"`
}

type collectionDataSource struct{ client *directus.Client }

func CollectionDataSource() datasource.DataSource { return &collectionDataSource{} }

func (d *collectionDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_collection"
}

func (d *collectionDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus collection by name.",
		Attributes: map[string]dschema.Attribute{
			"collection": dsRequiredString("Name of the collection to look up."),
			"meta": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"icon":             dsComputedString("Material icon name."),
					"note":             dsComputedString("Description."),
					"color":            dsComputedString("Accent color."),
					"display_template": dsComputedString("Display template."),
					"hidden":           dsComputedBool("Hidden in the app."),
					"singleton":        dsComputedBool("Singleton collection."),
					"sort_field":       dsComputedString("Manual sort field."),
					"group":            dsComputedString("Parent collection."),
					"collapse":         dsComputedString("Default collapse behavior."),
				},
			},
			"schema": dschema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]dschema.Attribute{
					"name":    dsComputedString("Table name."),
					"schema":  dsComputedString("Database schema."),
					"comment": dsComputedString("Table comment."),
				},
			},
		},
	}
}

func (d *collectionDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *collectionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config collectionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	col, err := d.client.GetCollectionByName(config.Collection.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus collection", err.Error())
		return
	}
	if col == nil {
		response.Diagnostics.AddError("Directus collection not found", "No collection named "+config.Collection.ValueString())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, collectionToDataSourceModel(col))...)
}

// collectionToDataSourceModel maps the client type into the data source model.
// Meta and schema are always-computed here, so they are populated whenever the
// server returns them.
func collectionToDataSourceModel(col *directus.Collection) collectionDataSourceModel {
	model := collectionDataSourceModel{
		Collection: types.StringValue(col.Collection),
	}
	if col.Meta != nil {
		model.Meta = &collectionDataSourceMetaModel{
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
	if col.Schema != nil {
		model.Schema = &collectionDataSourceSchemaModel{
			Name:    types.StringValue(col.Schema.Name),
			Schema:  types.StringValue(col.Schema.Schema),
			Comment: types.StringValue(col.Schema.Comment),
		}
	}
	return model
}
