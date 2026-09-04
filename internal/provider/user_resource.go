// Copyright IBM Corp. 2026

package provider

import (
	"context"

	"github.com/chop-sticks/directus-client-go/directus"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &userResource{}
	_ resource.ResourceWithConfigure   = &userResource{}
	_ resource.ResourceWithImportState = &userResource{}
)

func UserResource() resource.Resource { return &userResource{} }

type userResource struct {
	client *directus.Client
}

type userResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Email              types.String `tfsdk:"email"`
	Password           types.String `tfsdk:"password"`
	FirstName          types.String `tfsdk:"first_name"`
	LastName           types.String `tfsdk:"last_name"`
	Status             types.String `tfsdk:"status"`
	Role               types.String `tfsdk:"role"`
	Title              types.String `tfsdk:"title"`
	Description        types.String `tfsdk:"description"`
	Location           types.String `tfsdk:"location"`
	Language           types.String `tfsdk:"language"`
	Appearance         types.String `tfsdk:"appearance"`
	ThemeLight         types.String `tfsdk:"theme_light"`
	ThemeDark          types.String `tfsdk:"theme_dark"`
	Tags               types.List   `tfsdk:"tags"`
	EmailNotifications types.Bool   `tfsdk:"email_notifications"`
	ExternalIdentifier types.String `tfsdk:"external_identifier"`
	Provider           types.String `tfsdk:"auth_provider"`
}

func (r *userResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_user"
}

func (r *userResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "Manages a Directus user (directus_users).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Server-assigned user UUID.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address (login identifier).",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "User password. Write-only; never read back from the API.",
				Optional:            true,
				Sensitive:           true,
			},
			"first_name":          optionalComputedString("First name."),
			"last_name":           optionalComputedString("Last name."),
			"status":              optionalComputedString("Account status (active, invited, draft, suspended, archived)."),
			"role":                optionalComputedString("UUID of the user's role."),
			"title":               optionalComputedString("Job title."),
			"description":         optionalComputedString("Description/bio."),
			"location":            optionalComputedString("Location."),
			"language":            optionalComputedString("Preferred language code."),
			"appearance":          optionalComputedString("Appearance preference (auto, light, dark)."),
			"theme_light":         optionalComputedString("Light theme name."),
			"theme_dark":          optionalComputedString("Dark theme name."),
			"external_identifier": optionalComputedString("External identifier for SSO."),
			"auth_provider":       computedString("Auth provider that owns this user."),
			"tags":                optionalComputedStringList("Freeform tags on the user."),
			"email_notifications": optionalComputedBool("Whether the user receives email notifications."),
		},
	}
}

func (r *userResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan userResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	created, err := r.client.CreateUser(userModelToClient(ctx, plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error creating Directus user", err.Error())
		return
	}

	model := userToModel(ctx, created, &response.Diagnostics)
	model.Password = plan.Password
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *userResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state userResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(state.ID.ValueString(), nil)
	if err != nil {
		if isNotFound(err) {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("Error reading Directus user", err.Error())
		return
	}
	if user == nil {
		response.State.RemoveResource(ctx)
		return
	}

	model := userToModel(ctx, user, &response.Diagnostics)
	// Password is write-only; the API never returns it, so preserve state.
	model.Password = state.Password
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *userResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan userResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	updated, err := r.client.PatchUser(plan.ID.ValueString(), userModelToClient(ctx, plan, &response.Diagnostics), nil)
	if response.Diagnostics.HasError() {
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Error updating Directus user", err.Error())
		return
	}

	model := userToModel(ctx, updated, &response.Diagnostics)
	model.Password = plan.Password
	response.Diagnostics.Append(response.State.Set(ctx, model)...)
}

func (r *userResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state userResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteUser(state.ID.ValueString()); err != nil {
		if isNotFound(err) {
			return
		}
		response.Diagnostics.AddError("Error deleting Directus user", err.Error())
	}
}

func (r *userResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.client = configureClient(request.ProviderData, &response.Diagnostics)
}

func (r *userResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

// --- model <-> client mapping ---

func userModelToClient(ctx context.Context, m userResourceModel, diags *diag.Diagnostics) *directus.User {
	user := &directus.User{
		Email:              m.Email.ValueString(),
		Password:           m.Password.ValueString(),
		FirstName:          m.FirstName.ValueString(),
		LastName:           m.LastName.ValueString(),
		Status:             m.Status.ValueString(),
		Title:              m.Title.ValueString(),
		Description:        m.Description.ValueString(),
		Location:           m.Location.ValueString(),
		Language:           m.Language.ValueString(),
		Appearance:         m.Appearance.ValueString(),
		ThemeLight:         m.ThemeLight.ValueString(),
		ThemeDark:          m.ThemeDark.ValueString(),
		ExternalIdentifier: m.ExternalIdentifier.ValueString(),
		EmailNotifications: m.EmailNotifications.ValueBool(),
	}
	if !m.Role.IsNull() && !m.Role.IsUnknown() {
		user.Role = m.Role.ValueString()
	}
	if !m.Tags.IsNull() && !m.Tags.IsUnknown() {
		diags.Append(m.Tags.ElementsAs(ctx, &user.Tags, false)...)
	}
	return user
}

func userToModel(ctx context.Context, u *directus.User, diags *diag.Diagnostics) userResourceModel {
	tags, d := types.ListValueFrom(ctx, types.StringType, u.Tags)
	diags.Append(d...)
	return userResourceModel{
		ID:                 types.StringValue(u.ID),
		Email:              types.StringValue(u.Email),
		FirstName:          types.StringValue(u.FirstName),
		LastName:           types.StringValue(u.LastName),
		Status:             types.StringValue(u.Status),
		Role:               anyToStringID(u.Role),
		Title:              types.StringValue(u.Title),
		Description:        types.StringValue(u.Description),
		Location:           types.StringValue(u.Location),
		Language:           types.StringValue(u.Language),
		Appearance:         types.StringValue(u.Appearance),
		ThemeLight:         types.StringValue(u.ThemeLight),
		ThemeDark:          types.StringValue(u.ThemeDark),
		Tags:               tags,
		EmailNotifications: types.BoolValue(u.EmailNotifications),
		ExternalIdentifier: types.StringValue(u.ExternalIdentifier),
		Provider:           types.StringValue(u.Provider),
	}
}
