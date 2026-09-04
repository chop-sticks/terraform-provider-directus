// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type userDataSource struct{ client *directus.Client }

func UserDataSource() datasource.DataSource { return &userDataSource{} }

func (d *userDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_user"
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads a Directus user by id.",
		Attributes: map[string]dschema.Attribute{
			"id":                  dsRequiredString("UUID of the user to look up."),
			"email":               dsComputedString("Email address."),
			"password":            dschema.StringAttribute{MarkdownDescription: "Always null; passwords are never returned.", Computed: true, Sensitive: true},
			"first_name":          dsComputedString("First name."),
			"last_name":           dsComputedString("Last name."),
			"status":              dsComputedString("Account status."),
			"role":                dsComputedString("Role UUID."),
			"title":               dsComputedString("Job title."),
			"description":         dsComputedString("Description."),
			"location":            dsComputedString("Location."),
			"language":            dsComputedString("Preferred language."),
			"appearance":          dsComputedString("Appearance preference."),
			"theme_light":         dsComputedString("Light theme."),
			"theme_dark":          dsComputedString("Dark theme."),
			"tags":                dsComputedStringList("Freeform tags."),
			"email_notifications": dsComputedBool("Email notifications enabled."),
			"external_identifier": dsComputedString("External identifier."),
			"auth_provider":       dsComputedString("Auth provider."),
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *userDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var config userResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	user, err := d.client.GetUser(config.ID.ValueString(), nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus user", err.Error())
		return
	}
	if user == nil {
		response.Diagnostics.AddError("Directus user not found", "No user with id "+config.ID.ValueString())
		return
	}

	model := userToModel(ctx, user, &response.Diagnostics)
	model.Password = types.StringNull()
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}
