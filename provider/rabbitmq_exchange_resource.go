package provider

import (
	"context"
	"fmt"
	"strings"

	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &RabbitmqExchangeResource{}

func NewRabbitmqExchangeResource() resource.Resource {
	return &RabbitmqExchangeResource{}
}

type RabbitmqExchangeResource struct {
	providerData *RabbitmqProviderData
}

type RabbitmqExchangeSettingsModel struct {
	Type       types.String `tfsdk:"type"`
	Durable    types.Bool   `tfsdk:"durable"`
	AutoDelete types.Bool   `tfsdk:"auto_delete"`
	Arguments  types.Map    `tfsdk:"arguments"`
}

type RabbitmqExchangeResourceModel struct {
	Name     types.String                   `tfsdk:"name"`
	Vhost    types.String                   `tfsdk:"vhost"`
	Settings *RabbitmqExchangeSettingsModel `tfsdk:"settings"`
	Id       types.String                   `tfsdk:"id"`
}

func (r *RabbitmqExchangeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.providerData = req.ProviderData.(*RabbitmqProviderData)
}

func (r *RabbitmqExchangeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_exchange"
}

func (r *RabbitmqExchangeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vhost": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"settings": schema.SingleNestedAttribute{
				Required: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"durable": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"auto_delete": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"arguments": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
						Computed:    true,
						PlanModifiers: []planmodifier.Map{
							mapplanmodifier.RequiresReplace(),
						},
					},
				},
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

func (r *RabbitmqExchangeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name@vhost. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vhost"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}

func (r *RabbitmqExchangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RabbitmqExchangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	vhost := "/"
	if !plan.Vhost.IsNull() {
		vhost = plan.Vhost.ValueString()
	}

	id := fmt.Sprintf("%s@%s", name, vhost)
	plan.Id = types.StringValue(id)
	plan.Vhost = types.StringValue(vhost)

	if plan.Settings.Durable.IsNull() {
		plan.Settings.Durable = types.BoolValue(false)
	}

	if plan.Settings.AutoDelete.IsNull() {
		plan.Settings.AutoDelete = types.BoolValue(false)
	}

	if plan.Settings.Arguments.IsNull() {
		plan.Settings.Arguments = types.MapNull(types.StringType)
	}

	tflog.Trace(ctx, "creating rabbitmq exchange", map[string]interface{}{
		"name":  name,
		"vhost": vhost,
	})

	arguments := make(map[string]interface{})
	if !plan.Settings.Arguments.IsNull() {
		for k, v := range plan.Settings.Arguments.Elements() {
			arguments[k] = v.String()
		}
	}

	exchangeSettings := rabbithole.ExchangeSettings{
		Type:       plan.Settings.Type.ValueString(),
		Durable:    plan.Settings.Durable.ValueBool(),
		AutoDelete: plan.Settings.AutoDelete.ValueBool(),
		Arguments:  arguments,
	}

	rmqc := r.providerData.rabbitmqClient
	response, err := rmqc.DeclareExchange(vhost, name, exchangeSettings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating RabbitMQ exchange",
			fmt.Sprintf("Could not create RabbitMQ exchange %s in vhost %s: %s", name, vhost, err.Error()),
		)
		return
	}

	if response.StatusCode >= 400 {
		resp.Diagnostics.AddError(
			"Error creating RabbitMQ exchange",
			fmt.Sprintf("Could not create RabbitMQ exchange %s in vhost %s: %s", name, vhost, response.Status),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RabbitmqExchangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RabbitmqExchangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	vhost := state.Vhost.ValueString()

	tflog.Trace(ctx, "reading rabbitmq exchange", map[string]interface{}{
		"name":  name,
		"vhost": vhost,
	})

	rmqc := r.providerData.rabbitmqClient
	exchange, err := rmqc.GetExchange(vhost, name)
	if err != nil {
		if rabbitErr, ok := err.(rabbithole.ErrorResponse); ok && rabbitErr.StatusCode == 404 {
			tflog.Warn(ctx, "rabbitmq exchange not found, removing from state", map[string]interface{}{
				"name":  name,
				"vhost": vhost,
			})
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading RabbitMQ exchange",
			fmt.Sprintf("Could not read RabbitMQ exchange %s in vhost %s: %s", name, vhost, err.Error()),
		)
		return
	}

	if exchange == nil {
		tflog.Warn(ctx, "rabbitmq exchange not found, removing from state", map[string]interface{}{
			"name":  name,
			"vhost": vhost,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(exchange.Name)
	state.Vhost = types.StringValue(exchange.Vhost)

	if state.Settings == nil {
		state.Settings = &RabbitmqExchangeSettingsModel{}
	}
	state.Settings.Type = types.StringValue(exchange.Type)
	state.Settings.Durable = types.BoolValue(exchange.Durable)
	state.Settings.AutoDelete = types.BoolValue(exchange.AutoDelete)

	if exchange.Arguments == nil {
		state.Settings.Arguments = types.MapNull(types.StringType)
	} else {
		args, diags := types.MapValueFrom(ctx, types.StringType, exchange.Arguments)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Settings.Arguments = args
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RabbitmqExchangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes are RequiresReplace, so this function should not be called.
}

func (r *RabbitmqExchangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RabbitmqExchangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	vhost := state.Vhost.ValueString()

	tflog.Trace(ctx, "deleting rabbitmq exchange", map[string]interface{}{
		"name":  name,
		"vhost": vhost,
	})

	rmqc := r.providerData.rabbitmqClient
	response, err := rmqc.DeleteExchange(vhost, name)
	if err != nil {
		if rabbitErr, ok := err.(rabbithole.ErrorResponse); ok && rabbitErr.StatusCode == 404 {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Exchange",
			fmt.Sprintf("Could not delete RabbitMQ exchange %s in vhost %s: %s", name, vhost, err.Error()),
		)
		return
	}

	if response.StatusCode >= 400 && response.StatusCode != 404 {
		resp.Diagnostics.AddError(
			"Error Deleting RabbitMQ Exchange",
			fmt.Sprintf("Could not delete RabbitMQ exchange %s in vhost %s: %s", name, vhost, response.Status),
		)
		return
	}
}
