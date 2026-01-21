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

var _ resource.Resource = &RabbitmqPermissionsResource{}

func NewRabbitmqPermissionsResource() resource.Resource {
	return &RabbitmqPermissionsResource{}
}

type RabbitmqPermissionsResource struct {
	providerData *RabbitmqProviderData
}

type RabbitmqPermissionsResourceModel struct {
	User      types.String `tfsdk:"user"`
	Vhost     types.String `tfsdk:"vhost"`
	Configure types.String `tfsdk:"configure"`
	Write     types.String `tfsdk:"write"`
	Read      types.String `tfsdk:"read"`
	Id        types.String `tfsdk:"id"`
}

func (r *RabbitmqPermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerData = req.ProviderData.(*RabbitmqProviderData)
}

func (r *RabbitmqPermissionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permissions"
}

func (r *RabbitmqPermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"configure": schema.StringAttribute{
				Required:    true,
				Description: "The configure permissions.",
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

func (r *RabbitmqPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user@vhost. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vhost"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *RabbitmqPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RabbitmqPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := plan.User.ValueString()
	vhost := "/"
	if !plan.Vhost.IsNull() {
		vhost = plan.Vhost.ValueString()
	}

	id := fmt.Sprintf("%s@%s", user, vhost)
	plan.Id = types.StringValue(id)
	plan.Vhost = types.StringValue(vhost)

	tflog.Trace(ctx, "creating rabbitmq permissions", map[string]interface{}{
		"user":  user,
		"vhost": vhost,
	})

	err := r.setPermissions(user, vhost, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating RabbitMQ Permissions",
			fmt.Sprintf("Could not create RabbitMQ permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RabbitmqPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := state.User.ValueString()
	vhost := state.Vhost.ValueString()

	tflog.Trace(ctx, "reading rabbitmq permissions", map[string]interface{}{
		"user":  user,
		"vhost": vhost,
	})

	permissions, err := r.providerData.rabbitmqClient.GetPermissionsIn(vhost, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ Permissions",
			fmt.Sprintf("Could not read RabbitMQ permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	state.Configure = types.StringValue(permissions.Configure)
	state.Write = types.StringValue(permissions.Write)
	state.Read = types.StringValue(permissions.Read)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RabbitmqPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RabbitmqPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := plan.User.ValueString()
	vhost := plan.Vhost.ValueString()

	tflog.Trace(ctx, "updating rabbitmq permissions", map[string]interface{}{
		"user":  user,
		"vhost": vhost,
	})

	err := r.setPermissions(user, vhost, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating RabbitMQ Permissions",
			fmt.Sprintf("Could not update RabbitMQ permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RabbitmqPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := state.User.ValueString()
	vhost := state.Vhost.ValueString()

	tflog.Trace(ctx, "deleting rabbitmq permissions", map[string]interface{}{
		"user":  user,
		"vhost": vhost,
	})

	response, err := r.providerData.rabbitmqClient.ClearPermissionsIn(vhost, user)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Permissions",
			fmt.Sprintf("Could not delete RabbitMQ permissions for user %s in vhost %s: %s", user, vhost, err.Error()),
		)
		return
	}

	if response.StatusCode == 404 {
		return
	}

	if response.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Permissions",
			fmt.Sprintf("Could not delete RabbitMQ permissions for user %s in vhost %s: %s", user, vhost, response.Status),
		)
		return
	}
}

func (r *RabbitmqPermissionsResource) setPermissions(user, vhost string, plan *RabbitmqPermissionsResourceModel) error {
	permissions := rabbithole.Permissions{
		Configure: plan.Configure.ValueString(),
		Write:     plan.Write.ValueString(),
		Read:      plan.Read.ValueString(),
	}

	response, err := r.providerData.rabbitmqClient.UpdatePermissionsIn(vhost, user, permissions)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("error setting permissions: %s", response.Status)
	}

	return nil
}
