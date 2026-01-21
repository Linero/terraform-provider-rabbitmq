package provider

import (
	"context"
	"fmt"
	"reflect"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RabbitmqUserResource{}

func NewRabbitmqUserResource() resource.Resource {
	return &RabbitmqUserResource{}
}

type RabbitmqUserResource struct {
	providerData *RabbitmqProviderData
}

type RabbitmqUserResourceModel struct {
	Name              types.String `tfsdk:"name"`
	PasswordWo        types.String `tfsdk:"password_wo"`
	PasswordWoVersion types.String `tfsdk:"password_wo_version"`
	Tags              types.List   `tfsdk:"tags"`
}

func (r *RabbitmqUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerData = req.ProviderData.(*RabbitmqProviderData)
}

func (r *RabbitmqUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *RabbitmqUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the user.",
			},
			"password_wo": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				WriteOnly:   true,
				Description: "Write-only password for the user.",
			},
			"password_wo_version": schema.StringAttribute{
				Required:    true,
				Description: "Version string for password. Changing this value forces a password update even if password_wo hasn't changed in the configuration. Use this to rotate passwords.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RabbitmqUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

func (r *RabbitmqUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan, config RabbitmqUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	password := config.PasswordWo.ValueString()
	tags := ListToStringArray(plan.Tags)

	tflog.Trace(ctx, "creating rabbitmq user", map[string]interface{}{
		"user": plan.Name.ValueString(),
	})
	err := r.CreateUser(name, password, tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating RabbitMQ User",
			fmt.Sprintf("Could not create RabbitMQ user %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RabbitmqUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	tflog.Trace(ctx, "reading rabbitmq user", map[string]interface{}{
		"user": name,
	})
	user, err := r.ReadUser(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ User",
			fmt.Sprintf("Could not read RabbitMQ user %s: %s", name, err.Error()),
		)
		return
	}

	r.LoadUserIntoState(&state, user)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RabbitmqUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config RabbitmqUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	tags := ListToStringArray(plan.Tags)
	password := ""

	if plan.PasswordWoVersion.ValueString() != state.PasswordWoVersion.ValueString() {
		password = config.PasswordWo.ValueString()
	}

	tflog.Trace(ctx, "reading rabbitmq user", map[string]interface{}{
		"user": name,
	})
	user, err := r.ReadUser(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ User",
			fmt.Sprintf("Could not read RabbitMQ user %s: %s", name, err.Error()),
		)
		return
	}

	tflog.Trace(ctx, "updating rabbitmq user", map[string]interface{}{
		"user": name,
	})

	err = r.UpdateUser(user, password, tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating RabbitMQ User",
			fmt.Sprintf("Could not update RabbitMQ user %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RabbitmqUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	name := state.Name.ValueString()
	tflog.Trace(ctx, "deleting rabbitmq user", map[string]interface{}{
		"user": name,
	})
	err := r.DeleteUser(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ User",
			fmt.Sprintf("Could not delete RabbitMQ user %s: %s", name, err.Error()),
		)
		return
	}
}

func (r *RabbitmqUserResource) CreateUser(name string, password string, tags []string) error {
	rmqc := r.providerData.rabbitmqClient

	userSettings := rabbithole.UserSettings{
		Password: password,
		Tags:     tags,
	}

	resp, err := rmqc.PutUser(name, userSettings)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error creating RabbitMQ user: %s", resp.Status)
	}

	return nil
}

func (r *RabbitmqUserResource) ReadUser(name string) (*rabbithole.UserInfo, error) {
	rmqc := r.providerData.rabbitmqClient

	user, err := rmqc.GetUser(name)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("user %s not found", name)
	}

	return user, nil
}

func (r *RabbitmqUserResource) LoadUserIntoState(model *RabbitmqUserResourceModel, user *rabbithole.UserInfo) {
	model.Name = types.StringValue(user.Name)
	var tagList []attr.Value
	for _, v := range user.Tags {
		if v != "" {
			tagList = append(tagList, types.StringValue(v))
		}
	}
	if len(tagList) > 0 {
		model.Tags = types.ListValueMust(types.StringType, tagList)
	} else {
		model.Tags = types.ListValueMust(types.StringType, []attr.Value{})
	}
}

func (r *RabbitmqUserResource) UpdateUser(user *rabbithole.UserInfo, newPassword string, newTags []string) error {
	rmqc := r.providerData.rabbitmqClient

	userSettings := rabbithole.UserSettings{
		PasswordHash:     user.PasswordHash,
		HashingAlgorithm: user.HashingAlgorithm,
		Tags:             user.Tags,
	}

	if newPassword != "" {

		hash := rabbithole.Base64EncodedSaltedPasswordHashSHA256(newPassword)
		userSettings.PasswordHash = hash
	}

	equalTags := reflect.DeepEqual(user.Tags, newTags)

	if !equalTags {
		userSettings.Tags = newTags

	}

	resp, err := rmqc.PutUser(user.Name, userSettings)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error updating RabbitMQ user: %s", resp.Status)
	}

	return nil
}

func (r *RabbitmqUserResource) DeleteUser(name string) error {
	rmqc := r.providerData.rabbitmqClient

	resp, err := rmqc.DeleteUser(name)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("error deleting RabbitMQ user: %s", resp.Status)
	}

	return nil
}

func ListToStringArray(list types.List) (array []string) {
	elements := list.Elements()
	for _, v := range elements {
		array = append(array, v.(types.String).ValueString())
	}
	return array
}
