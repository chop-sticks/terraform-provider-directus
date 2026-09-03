// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type policyDataSource struct{ client *directus.Client }

func PolicyDataSource() datasource.DataSource { return &policyDataSource{} }

func (d *policyDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_policy"
}

func (d *policyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus access policy by id.",
		Attributes: map[string]dschema.Attribute{
			"id":           dsRequiredString("UUID of the policy to look up."),
			"name":         dsComputedString("Name of the policy."),
			"icon":         dsComputedString("Material icon name."),
			"description":  dsComputedString("Description."),
			"ip_access":    dsComputedString("Allowed IP list."),
			"enforce_tfa":  dsComputedBool("Whether TFA is enforced."),
			"admin_access": dsComputedBool("Whether full admin access is granted."),
			"app_access":   dsComputedBool("Whether app access is granted."),
		},
	}
}

func (d *policyDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *policyDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config policyResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	policy, err := d.client.GetPolicy(config.ID.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus policy", err.Error())
		return
	}
	if policy == nil {
		response.Diagnostics.AddError("Directus policy not found", "No policy with id "+config.ID.ValueString())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, policyToModel(policy))...)
}
