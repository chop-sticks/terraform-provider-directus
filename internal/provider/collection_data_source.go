// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

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
	var config collectionResourceModel
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

	response.Diagnostics.Append(response.State.Set(ctx, collectionToModel(col, true, true))...)
}
