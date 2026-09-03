// Copyright IBM Corp. 2026

package provider

import (
	"context"
	"os"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DirectusProvider{}

type DirectusProvider struct {
	version string
}

type DirectusProviderModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

func (d *DirectusProvider) Metadata(ctx context.Context, request provider.MetadataRequest, response *provider.MetadataResponse) {
	response.TypeName = "directus"
	response.Version = d.version
}

func (d *DirectusProvider) Schema(ctx context.Context, request provider.SchemaRequest, response *provider.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func (d *DirectusProvider) Configure(ctx context.Context, request provider.ConfigureRequest, response *provider.ConfigureResponse) {
	var config DirectusProviderModel
	diags := request.Config.Get(ctx, &config)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.URL.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown Directus URL",
			"Directus URL must be known to configure the provider.")
	}
	url := os.Getenv("DIRECTUS_URL")
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}
	if url == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Mising Directus URL",
			"The provider cannot create the Directus API client as there is a missing or empty value for the Directus URL. "+
				"Set the url value in the configuration or use the DIRECTUS_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
		return
	}

	if config.Token.IsUnknown() {
		response.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown Directus Token",
			"Directus Token must be known to configure the provider.")
	}
	token := os.Getenv("DIRECTUS_TOKEN")
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if token == "" {
		response.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Directus Token",
			"The provider cannot create the Directus API client as there is a missing or empty value for the Directus Token. "+
				"Set the token value in the configuration or use the DIRECTUS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	client, err := directus.NewClient(&url, &token)
	if err != nil {
		response.Diagnostics.AddError(
			"Failed to create Directus API client",
			"The provider cannot create the Directus API client. "+
				"Ensure the URL and Token are valid and correctly configured. "+
				"Check the Directus documentation for more information."+
				"Directus API client Error: "+err.Error(),
		)
		return
	}

	response.DataSourceData = client
	response.ResourceData = client
}

func (d *DirectusProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		CollectionDataSource,
		RoleDataSource,
		PolicyDataSource,
		FolderDataSource,
		UserDataSource,
		SettingsDataSource,
		ServerInfoDataSource,
	}
}

func (d *DirectusProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		CollectionResource,
		RoleResource,
		PolicyResource,
		FolderResource,
		TranslationResource,
		PresetResource,
		DashboardResource,
		PanelResource,
		FlowResource,
		OperationResource,
		FieldResource,
		RelationResource,
		PermissionResource,
		SettingsResource,
		UserResource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DirectusProvider{
			version: version,
		}
	}
}
