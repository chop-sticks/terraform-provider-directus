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
	_ resource.Resource                = &folderResource{}
	_ resource.ResourceWithConfigure   = &folderResource{}
	_ resource.ResourceWithImportState = &folderResource{}
)

func FolderResource() resource.Resource { return &folderResource{} }

type folderResource struct {
	client *directus.Client
}

type folderResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Parent types.String `tfsdk:"parent"`
}

func (r *folderResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_folder"
}

func (r *folderResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	// Directus exposes no PATCH /folders endpoint, so every mutable attribute
	// forces replacement; Update is therefore never invoked.
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus file folder (directus_folders). Directus has no folder-update endpoint, so any change replaces the folder.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned folder UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the folder. Changing this forces a new resource.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"parent": schema.StringAttribute{
				MarkdownDescription: "UUID of the parent folder. Changing this forces a new resource.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *folderResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan folderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	folder := &directus.Folder{Name: plan.Name.ValueString()}
	if !plan.Parent.IsNull() && !plan.Parent.IsUnknown() {
		folder.Parent = plan.Parent.ValueString()
	}

	created, err := r.client.CreateFolder(folder, nil)
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus folder", err.Error())
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, folderToModel(created))...)
}

func (r *folderResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state folderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	folder, err := r.client.GetFolder(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus folder", err.Error())
		return
	}
	if folder == nil {
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, folderToModel(folder))...)
}

// Update is unreachable: all mutable attributes force replacement. It exists to
// satisfy the resource.Resource interface and simply persists the plan.
func (r *folderResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan folderResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, plan)...)
}

func (r *folderResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state folderResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteFolder(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus folder", err.Error())
	}
}

func (r *folderResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *folderResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func folderToModel(folder *directus.Folder) folderResourceModel {
	return folderResourceModel{
		ID:     types.StringValue(folder.ID),
		Name:   types.StringValue(folder.Name),
		Parent: anyToStringID(folder.Parent),
	}
}
