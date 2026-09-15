// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type folderDataSource struct{ client *directus.Client }

func FolderDataSource() datasource.DataSource { return &folderDataSource{} }

func (d *folderDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_folder"
}

func (d *folderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus folder by id.",
		Attributes: map[string]dschema.Attribute{
			"id":     dsRequiredString("UUID of the folder to look up."),
			"name":   dsComputedString("Name of the folder."),
			"parent": dsComputedString("Parent folder UUID."),
		},
	}
}

func (d *folderDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *folderDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config folderResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	folder, err := d.client.GetFolder(config.ID.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus folder", err.Error())
		return
	}
	if folder == nil {
		response.Diagnostics.AddError("Directus folder not found", "No folder with id "+config.ID.ValueString())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, folderToModel(folder))...)
}
