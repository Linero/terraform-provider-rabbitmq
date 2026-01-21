package provider

import (
	"context"
	"fmt"
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RabbitmqTopicPermissionsResource{}

func NewRabbitmqTopicPermissionsResource() resource.Resource {
	return &RabbitmqTopicPermissionsResource{}
}

type RabbitmqTopicPermissionsResource struct {
	providerData *RabbitmqProviderData
}

type RabbitmqTopicPermissionsResourceModel struct {
	User     types.String `tfsdk:"user"`
	Vhost    types.String `tfsdk:"vhost"`
	Exchange types.String `tfsdk:"exchange"`
	Write    types.String `tfsdk:"write"`
	Read     types.String `tfsdk:"read"`
	Id       types.String `tfsdk:"id"`
}

func (r *RabbitmqTopicPermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerData = req.ProviderData.(*RabbitmqProviderData)
}

func (r *RabbitmqTopicPermissionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_topic_permissions"
}

func (r *RabbitmqTopicPermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Required:    true,
				Description: "The user to grant permissions to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vhost": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The vhost to grant permissions for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"exchange": schema.StringAttribute{
				Required:    true,
				Description: "The exchange to apply permissions to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"write": schema.StringAttribute{
				Required:    true,
				Description: "The write permissions.",
			},
			"read": schema.StringAttribute{
				Required:    true,
				Description: "The read permissions.",
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *RabbitmqTopicPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user@vhost@exchange. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vhost"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("exchange"), parts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *RabbitmqTopicPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RabbitmqTopicPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := plan.User.ValueString()
	vhost := "/"
	if !plan.Vhost.IsNull() {
		vhost = plan.Vhost.ValueString()
	}
	exchange := plan.Exchange.ValueString()

	id := fmt.Sprintf("%s@%s@%s", user, vhost, exchange)
	plan.Id = types.StringValue(id)
	plan.Vhost = types.StringValue(vhost)

	tflog.Trace(ctx, "creating rabbitmq topic permissions", map[string]interface{}{
		"user":     user,
		"vhost":    vhost,
		"exchange": exchange,
	})

	err := r.setTopicPermissions(user, vhost, exchange, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating RabbitMQ Topic Permissions",
			fmt.Sprintf("Could not create RabbitMQ topic permissions for user %s in vhost %s on exchange %s: %s", user, vhost, exchange, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqTopicPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RabbitmqTopicPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := state.User.ValueString()
	vhost := state.Vhost.ValueString()
	exchange := state.Exchange.ValueString()

	tflog.Trace(ctx, "reading rabbitmq topic permissions", map[string]interface{}{
		"user":     user,
		"vhost":    vhost,
		"exchange": exchange,
	})

	permissions, err := r.providerData.rabbitmqClient.GetTopicPermissionsIn(vhost, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ Topic Permissions",
			fmt.Sprintf("Could not read RabbitMQ topic permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	var found *rabbithole.TopicPermissionInfo
	for _, p := range permissions {
		if p.Exchange == exchange {
			found = &p
			break
		}
	}

	if found == nil {
		// Topic permission does not exist
		resp.State.RemoveResource(ctx)
		return
	}

	state.Write = types.StringValue(found.Write)
	state.Read = types.StringValue(found.Read)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RabbitmqTopicPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RabbitmqTopicPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := plan.User.ValueString()
	vhost := plan.Vhost.ValueString()
	exchange := plan.Exchange.ValueString()

	tflog.Trace(ctx, "updating rabbitmq topic permissions", map[string]interface{}{
		"user":     user,
		"vhost":    vhost,
		"exchange": exchange,
	})

	err := r.setTopicPermissions(user, vhost, exchange, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating RabbitMQ Topic Permissions",
			fmt.Sprintf("Could not update RabbitMQ topic permissions for user %s in vhost %s on exchange %s: %s", user, vhost, exchange, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqTopicPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RabbitmqTopicPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := state.User.ValueString()
	vhost := state.Vhost.ValueString()
	exchange := state.Exchange.ValueString()

	tflog.Trace(ctx, "deleting rabbitmq topic permissions", map[string]interface{}{
		"user":     user,
		"vhost":    vhost,
		"exchange": exchange,
	})

	response, err := r.providerData.rabbitmqClient.ClearTopicPermissionsIn(vhost, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Topic Permissions",
			fmt.Sprintf("Could not delete RabbitMQ topic permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	if response.StatusCode == 404 {
		// The permissions were already deleted
		return
	}

	if response.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Topic Permissions",
			fmt.Sprintf("Could not delete RabbitMQ topic permissions for user %s in vhost %s on exchange %s: %s", user, vhost, exchange, response.Status),
		)
		return
	}
}

func (r *RabbitmqTopicPermissionsResource) setTopicPermissions(user, vhost, exchange string, plan *RabbitmqTopicPermissionsResourceModel) error {
	permissions := rabbithole.TopicPermissions{
		Exchange: exchange,
		Write:    plan.Write.ValueString(),
		Read:     plan.Read.ValueString(),
	}

	response, err := r.providerData.rabbitmqClient.UpdateTopicPermissionsIn(vhost, user, permissions)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error setting topic permissions: %s", response.Status)
	}

	return nil
}
