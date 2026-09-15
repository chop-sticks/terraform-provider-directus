// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &settingsResource{}
	_ resource.ResourceWithConfigure   = &settingsResource{}
	_ resource.ResourceWithImportState = &settingsResource{}
)

func SettingsResource() resource.Resource { return &settingsResource{} }

type settingsResource struct {
	client *directus.Client
}

type settingsResourceModel struct {
	ID                            types.Int64          `tfsdk:"id"`
	ProjectName                   types.String         `tfsdk:"project_name"`
	ProjectDescriptor             types.String         `tfsdk:"project_descriptor"`
	ProjectURL                    types.String         `tfsdk:"project_url"`
	ReportErrorURL                types.String         `tfsdk:"report_error_url"`
	ReportBugURL                  types.String         `tfsdk:"report_bug_url"`
	DefaultLanguage               types.String         `tfsdk:"default_language"`
	DefaultAppearance             types.String         `tfsdk:"default_appearance"`
	ProjectColor                  types.String         `tfsdk:"project_color"`
	ProjectLogo                   types.String         `tfsdk:"project_logo"`
	PublicRegistration            types.Bool           `tfsdk:"public_registration"`
	PublicRegistrationVerifyEmail types.Bool           `tfsdk:"public_registration_verify_email"`
	StorageAssetTransform         types.String         `tfsdk:"storage_asset_transform"`
	CustomCSS                     types.String         `tfsdk:"custom_css"`
	ModuleBar                     jsontypes.Normalized `tfsdk:"module_bar"`
	MapboxKey                     types.String         `tfsdk:"mapbox_key"`
	DefaultThemeLight             types.String         `tfsdk:"default_theme_light"`
	DefaultThemeDark              types.String         `tfsdk:"default_theme_dark"`
}

func (r *settingsResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_settings"
}

func (r *settingsResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages the Directus project settings singleton (directus_settings). There is one settings row; `terraform destroy` only removes it from state and does not reset the settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Settings singleton id.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_name":                     optionalComputedString("Project name shown in the app."),
			"project_descriptor":               optionalComputedString("Short project descriptor."),
			"project_url":                      optionalComputedString("Project URL."),
			"report_error_url":                 optionalComputedString("URL template for reporting errors."),
			"report_bug_url":                   optionalComputedString("URL template for reporting bugs."),
			"default_language":                 optionalComputedString("Default language code (e.g. en-US)."),
			"default_appearance":               optionalComputedString("Default appearance (auto, light, dark)."),
			"project_color":                    optionalComputedString("Project brand color (hex)."),
			"project_logo":                     optionalComputedString("File UUID of the project logo."),
			"storage_asset_transform":          optionalComputedString("Asset transformation policy (all, none, presets)."),
			"custom_css":                       optionalComputedString("Custom CSS injected into the app."),
			"mapbox_key":                       optionalComputedString("Mapbox access token."),
			"default_theme_light":              optionalComputedString("Default light theme name."),
			"default_theme_dark":               optionalComputedString("Default dark theme name."),
			"public_registration":              optionalComputedBool("Whether public user registration is enabled."),
			"public_registration_verify_email": optionalComputedBool("Whether public registration requires email verification."),
			"module_bar":                       normalizedJSONAttribute("Module bar configuration as a JSON array."),
		},
	}
}

// Create maps to PATCH /settings since the singleton always exists.
func (r *settingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan settingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchSettings(settingsModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error configuring Directus settings", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, settingsToModel(updated))...)
}

func (r *settingsResource) Read(ctx context.Context, _ resource.ReadRequest, response *resource.ReadResponse) {
	settings, err := r.client.GetSettings(nil)
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus settings", err.Error())
		return
	}
	if settings == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, settingsToModel(settings))...)
}

func (r *settingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan settingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchSettings(settingsModelToClient(plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus settings", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, settingsToModel(updated))...)
}

// Delete is a no-op: the settings singleton cannot be deleted. Terraform simply
// forgets the resource.
func (r *settingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *settingsResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *settingsResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importInt64ID(ctx, request, response)
}

// --- model <-> client mapping ---

func settingsModelToClient(m settingsResourceModel, diags *diag.Diagnostics) *directus.Settings {
	settings := &directus.Settings{
		ProjectName:                   m.ProjectName.ValueString(),
		ProjectDescriptor:             m.ProjectDescriptor.ValueString(),
		ProjectURL:                    m.ProjectURL.ValueString(),
		ReportErrorURL:                m.ReportErrorURL.ValueString(),
		ReportBugURL:                  m.ReportBugURL.ValueString(),
		DefaultLanguage:               m.DefaultLanguage.ValueString(),
		DefaultAppearance:             m.DefaultAppearance.ValueString(),
		ProjectColor:                  m.ProjectColor.ValueString(),
		PublicRegistration:            m.PublicRegistration.ValueBool(),
		PublicRegistrationVerifyEmail: m.PublicRegistrationVerifyEmail.ValueBool(),
		StorageAssetTransform:         m.StorageAssetTransform.ValueString(),
		CustomCSS:                     m.CustomCSS.ValueString(),
		ModuleBar:                     normalizedToAny(m.ModuleBar, diags),
		MapboxKey:                     m.MapboxKey.ValueString(),
		DefaultThemeLight:             m.DefaultThemeLight.ValueString(),
		DefaultThemeDark:              m.DefaultThemeDark.ValueString(),
	}
	if !m.ProjectLogo.IsNull() && !m.ProjectLogo.IsUnknown() {
		settings.ProjectLogo = m.ProjectLogo.ValueString()
	}
	return settings
}

func settingsToModel(s *directus.Settings) settingsResourceModel {
	return settingsResourceModel{
		ID:                            types.Int64Value(int64(s.ID)),
		ProjectName:                   types.StringValue(s.ProjectName),
		ProjectDescriptor:             types.StringValue(s.ProjectDescriptor),
		ProjectURL:                    types.StringValue(s.ProjectURL),
		ReportErrorURL:                types.StringValue(s.ReportErrorURL),
		ReportBugURL:                  types.StringValue(s.ReportBugURL),
		DefaultLanguage:               types.StringValue(s.DefaultLanguage),
		DefaultAppearance:             types.StringValue(s.DefaultAppearance),
		ProjectColor:                  types.StringValue(s.ProjectColor),
		ProjectLogo:                   anyToStringID(s.ProjectLogo),
		PublicRegistration:            types.BoolValue(s.PublicRegistration),
		PublicRegistrationVerifyEmail: types.BoolValue(s.PublicRegistrationVerifyEmail),
		StorageAssetTransform:         types.StringValue(s.StorageAssetTransform),
		CustomCSS:                     types.StringValue(s.CustomCSS),
		ModuleBar:                     anyToNormalized(s.ModuleBar),
		MapboxKey:                     types.StringValue(s.MapboxKey),
		DefaultThemeLight:             types.StringValue(s.DefaultThemeLight),
		DefaultThemeDark:              types.StringValue(s.DefaultThemeDark),
	}
}
