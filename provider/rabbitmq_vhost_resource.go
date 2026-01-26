package provider

import (
	"context"
	"fmt"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RabbitmqVhostResource{}

func NewRabbitmqVhostResource() resource.Resource {
	return &RabbitmqVhostResource{}
}

type RabbitmqVhostResource struct {
	providerData *RabbitmqProviderData
}

type RabbitmqVhostResourceModel struct {
	Name             types.String `tfsdk:"name"`
	Id               types.String `tfsdk:"id"`
	Description      types.String `tfsdk:"description"`
	DefaultQueueType types.String `tfsdk:"default_queue_type"`
	Tracing          types.Bool   `tfsdk:"tracing"`
	MaxConnections   types.String `tfsdk:"max_connections"`
	MaxQueues        types.String `tfsdk:"max_queues"`
}

func (r *RabbitmqVhostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerData = req.ProviderData.(*RabbitmqProviderData)
}

func (r *RabbitmqVhostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vhost"
}

func (r *RabbitmqVhostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the vhost.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "The description of the vhost.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_queue_type": schema.StringAttribute{
				Optional:    true,
				Description: "The default queue type for the vhost.",
			},
			"tracing": schema.BoolAttribute{
				Optional:    true,
				Description: "The tracing setting for the vhost.",
			},
			"max_connections": schema.StringAttribute{
				Optional:    true,
				Description: "The max connections for the vhost.",
			},
			"max_queues": schema.StringAttribute{
				Optional:    true,
				Description: "The max queues for the vhost.",
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

func (r *RabbitmqVhostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *RabbitmqVhostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RabbitmqVhostResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()

	tflog.Trace(ctx, "creating rabbitmq vhost", map[string]interface{}{
		"name": name,
	})

	rmqc := r.providerData.rabbitmqClient
	response, err := rmqc.PutVhost(name, rabbithole.VhostSettings{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating RabbitMQ vhost",
			fmt.Sprintf("Could not create RabbitMQ vhost %s: %s", name, err.Error()),
		)
		return
	}

	if response.StatusCode != 201 && response.StatusCode != 204 {
		resp.Diagnostics.AddError(
			"Error creating RabbitMQ vhost",
			fmt.Sprintf("Could not create RabbitMQ vhost %s, got status code %d", name, response.StatusCode),
		)
		return
	}

	plan.Id = types.StringValue(name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqVhostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RabbitmqVhostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()

	tflog.Trace(ctx, "reading rabbitmq vhost", map[string]interface{}{
		"name": name,
	})

	rmqc := r.providerData.rabbitmqClient
	vhost, err := rmqc.GetVhost(name)
	if err != nil {
		if rabbitErr, ok := err.(rabbithole.ErrorResponse); ok && rabbitErr.StatusCode == 404 {
			tflog.Warn(ctx, "rabbitmq vhost not found, removing from state", map[string]interface{}{
				"name": name,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ vhost",
			fmt.Sprintf("Could not read RabbitMQ vhost %s: %s", name, err.Error()),
		)
		return
	}

	if vhost == nil {
		tflog.Warn(ctx, "rabbitmq vhost not found, removing from state", map[string]interface{}{
			"name": name,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(vhost.Name)
	state.Id = types.StringValue(vhost.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RabbitmqVhostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r *RabbitmqVhostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RabbitmqVhostResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	tflog.Trace(ctx, "deleting rabbitmq vhost", map[string]interface{}{
		"name": name,
	})

	rmqc := r.providerData.rabbitmqClient
	response, err := rmqc.DeleteVhost(name)
	if err != nil {
		if rabbitErr, ok := err.(rabbithole.ErrorResponse); ok && rabbitErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Vhost",
			fmt.Sprintf("Could not delete RabbitMQ vhost %s: %s", name, err.Error()),
		)
		return
	}

	if response.StatusCode >= 400 && response.StatusCode != 404 {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Vhost",
			fmt.Sprintf("Could not delete RabbitMQ vhost %s: %s", name, response.Status),
		)
		return
	}
}
