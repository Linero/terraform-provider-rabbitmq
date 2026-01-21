package provider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	rabbithole "github.com/michaelklishin/rabbit-hole/v3"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*RabbitmqProvider)(nil)
var _ provider.ProviderWithEphemeralResources = (*RabbitmqProvider)(nil)

type RabbitmqProvider struct{}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &RabbitmqProvider{}
	}
}

type RabbitmqProviderModel struct {
	Address        types.String `tfsdk:"address"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	Insecure       types.Bool   `tfsdk:"insecure"`
	CacertFile     types.String `tfsdk:"cacert_file"`
	ClientcertFile types.String `tfsdk:"clientcert_file"`
	ClientkeyFile  types.String `tfsdk:"clientkey_file"`
	Proxy          types.String `tfsdk:"proxy"`
}

type RabbitmqProviderData struct {
	rabbitmqClient *rabbithole.Client
}

func (p *RabbitmqProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Required: true,
			},
			"username": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				Optional: true,
			},
			"cacert_file": schema.StringAttribute{
				Optional: true,
			},
			"clientcert_file": schema.StringAttribute{
				Optional: true,
			},
			"clientkey_file": schema.StringAttribute{
				Optional: true,
			},
			"proxy": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}
func (p *RabbitmqProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data RabbitmqProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rabbitmqClient, err := configureRmqClient(&data)
	if err != nil {
		resp.Diagnostics.AddError("Failed to configure RabbitMQ client", err.Error())
		return
	}

	providerData := &RabbitmqProviderData{
		rabbitmqClient: rabbitmqClient,
	}

	resp.ResourceData = providerData
	resp.DataSourceData = providerData
	resp.EphemeralResourceData = providerData
}
func (p *RabbitmqProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rabbitmq"
}

func (p *RabbitmqProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *RabbitmqProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRabbitmqUserResource,
		NewRabbitmqPermissionsResource,
		NewRabbitmqTopicPermissionsResource,
	}
}

func (p *RabbitmqProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func configureRmqClient(model *RabbitmqProviderModel) (*rabbithole.Client, error) {

	var username = model.Username.ValueString()
	var password = model.Password.ValueString()
	var endpoint = model.Address.ValueString()
	var insecure = model.Insecure.ValueBool()
	var cacertFile = model.CacertFile.ValueString()
	var clientcertFile = model.ClientcertFile.ValueString()
	var clientkeyFile = model.ClientkeyFile.ValueString()
	var proxy = model.Proxy.ValueString()

	tlsConfig := &tls.Config{}
	if cacertFile != "" {
		caCert, err := os.ReadFile(cacertFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	if clientcertFile != "" && clientkeyFile != "" {
		clientPair, err := tls.LoadX509KeyPair(clientcertFile, clientkeyFile)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{clientPair}
	}
	if insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	var proxyURL *url.URL
	if proxy != "" {
		var err error
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", proxy, err)
		}
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy: func(req *http.Request) (*url.URL, error) {
			if proxyURL != nil {
				return proxyURL, nil
			}

			return http.ProxyFromEnvironment(req)
		},
	}

	rabbitmqClient, err := rabbithole.NewTLSClient(endpoint, username, password, transport)
	if err != nil {
		return nil, err
	}

	return rabbitmqClient, nil
}
