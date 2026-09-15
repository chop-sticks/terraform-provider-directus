// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type accessDataSource struct{ client *directus.Client }

func AccessDataSource() datasource.DataSource { return &accessDataSource{} }

func (d *accessDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_access"
}

func (d *accessDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus access record (directus_access) by id.",
		Attributes: map[string]dschema.Attribute{
			"id":     dsRequiredString("UUID of the access record to look up."),
			"policy": dsComputedString("UUID of the assigned policy."),
			"role":   dsComputedString("UUID of the role the policy is assigned to (null for user assignments)."),
			"user":   dsComputedString("UUID of the user the policy is assigned to (null for role assignments)."),
			"sort":   dsComputedInt64("Manual sort order of the assignment."),
		},
	}
}

func (d *accessDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *accessDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config accessResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	access, err := d.client.GetAccess(config.ID.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus access", err.Error())
		return
	}
	if access == nil {
		response.Diagnostics.AddError("Directus access not found", "No access record with id "+config.ID.ValueString())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, accessToModel(access))...)
}
