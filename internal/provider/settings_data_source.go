// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type settingsDataSource struct{ client *directus.Client }

func SettingsDataSource() datasource.DataSource { return &settingsDataSource{} }

func (d *settingsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_settings"
}

func (d *settingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads the Directus project settings singleton.",
		Attributes: map[string]dschema.Attribute{
			"id":                               dsComputedInt64("Settings id."),
			"project_name":                     dsComputedString("Project name."),
			"project_descriptor":               dsComputedString("Project descriptor."),
			"project_url":                      dsComputedString("Project URL."),
			"report_error_url":                 dsComputedString("Report error URL."),
			"report_bug_url":                   dsComputedString("Report bug URL."),
			"default_language":                 dsComputedString("Default language."),
			"default_appearance":               dsComputedString("Default appearance."),
			"project_color":                    dsComputedString("Project color."),
			"project_logo":                     dsComputedString("Project logo file UUID."),
			"storage_asset_transform":          dsComputedString("Asset transform policy."),
			"custom_css":                       dsComputedString("Custom CSS."),
			"mapbox_key":                       dsComputedString("Mapbox key."),
			"default_theme_light":              dsComputedString("Default light theme."),
			"default_theme_dark":               dsComputedString("Default dark theme."),
			"public_registration":              dsComputedBool("Public registration enabled."),
			"public_registration_verify_email": dsComputedBool("Public registration email verification."),
			"module_bar":                       dsComputedJSON("Module bar configuration."),
		},
	}
}

func (d *settingsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *settingsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	settings, err := d.client.GetSettings(nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus settings", err.Error())
		return
	}
	if settings == nil {
		response.Diagnostics.AddError("Directus settings not found", "The settings singleton returned no data.")
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, settingsToModel(settings))...)
}
