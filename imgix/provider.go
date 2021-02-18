package imgix

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	AccessKey  string
	ApiBaseUrl string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Imgix API key. Can also be sourced from IMGIX_API_KEY environment variable",
				DefaultFunc: schema.EnvDefaultFunc("IMGIX_API_KEY", nil),
			},
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			config := Config{
				AccessKey: d.Get("api_key").(string),
			}
			client, err := NewClient(config)
			return client, diag.FromErr(err)
		},
		ResourcesMap: map[string]*schema.Resource{
			"imgix_source": resourceImgixSource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"imgix_source": dataSourceImgixSource(),
		},
	}
}
