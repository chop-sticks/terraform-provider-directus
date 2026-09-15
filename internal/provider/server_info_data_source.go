// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

type serverInfoDataSource struct{ client *directus.Client }

func ServerInfoDataSource() datasource.DataSource { return &serverInfoDataSource{} }

type serverInfoModel struct {
	Info jsontypes.Normalized `tfsdk:"info"`
}

func (d *serverInfoDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_server_info"
}

func (d *serverInfoDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = dschema.Schema{
		MarkdownDescription: "Reads Directus server information (/server/info).",
		Attributes: map[string]dschema.Attribute{
			"info": dsComputedJSON("Full server info payload as a JSON object."),
		},
	}
}

func (d *serverInfoDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (d *serverInfoDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	info, err := d.client.ServerInfo()
	if err != nil {
		response.Diagnostics.AddError("Error reading Directus server info", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, serverInfoModel{Info: mapToNormalized(info)})...)
}
