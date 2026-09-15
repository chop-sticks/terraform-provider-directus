// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type roleDataSource struct{ client *directus.Client }

func RoleDataSource() datasource.DataSource { return &roleDataSource{} }

func (d *roleDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_role"
}

func (d *roleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus role by id.",
		Attributes: map[string]dschema.Attribute{
			"id":          dsRequiredString("UUID of the role to look up."),
			"name":        dsComputedString("Name of the role."),
			"icon":        dsComputedString("Material icon name."),
			"description": dsComputedString("Description."),
			"parent":      dsComputedString("Parent role UUID."),
		},
	}
}

func (d *roleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *roleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config roleResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	role, err := d.client.GetRole(config.ID.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus role", err.Error())
		return
	}
	if role == nil {
		response.Diagnostics.AddError("Directus role not found", "No role with id "+config.ID.ValueString())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, roleToModel(role))...)
}
