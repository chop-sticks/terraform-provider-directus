// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &translationResource{}
	_ resource.ResourceWithConfigure   = &translationResource{}
	_ resource.ResourceWithImportState = &translationResource{}
)

func TranslationResource() resource.Resource { return &translationResource{} }

type translationResource struct {
	client *directus.Client
}

type translationResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Language types.String `tfsdk:"language"`
	Key      types.String `tfsdk:"key"`
	Value    types.String `tfsdk:"value"`
}

func (r *translationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_translation"
}

func (r *translationResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus custom translation string (directus_translations).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned translation UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language code (e.g. en-US).",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "Translation key.",
				Required:            true,
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "Translated value.",
				Required:            true,
			},
		},
	}
}

func (r *translationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan translationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateTranslation(translationModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus translation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, translationToModel(created))...)
}

func (r *translationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state translationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	translation, err := r.client.GetTranslation(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus translation", err.Error())
		return
	}
	if translation == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, translationToModel(translation))...)
}

func (r *translationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan translationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchTranslation(plan.ID.ValueString(), translationModelToClient(plan), nil)
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus translation", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, translationToModel(updated))...)
}

func (r *translationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state translationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTranslation(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus translation", err.Error())
	}
}

func (r *translationResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *translationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func translationModelToClient(m translationResourceModel) *directus.Translation {
	return &directus.Translation{
		Language: m.Language.ValueString(),
		Key:      m.Key.ValueString(),
		Value:    m.Value.ValueString(),
	}
}

func translationToModel(t *directus.Translation) translationResourceModel {
	return translationResourceModel{
		ID:       types.StringValue(t.ID),
		Language: types.StringValue(t.Language),
		Key:      types.StringValue(t.Key),
		Value:    types.StringValue(t.Value),
	}
}
