package imgix

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Config struct {
	AccessKey string
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Imgix API key",
				DefaultFunc: schema.EnvDefaultFunc("IMGIX_API_KEY", nil),
			},
		},
		ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			config := Config{
				AccessKey: d.Get("api_key").(string),
			}
			return NewClient(config), nil
		},
		ResourcesMap: map[string]*schema.Resource{
			"imgix_source": resourceImgixSource(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"imgix_source": dataSourceImgixSource(),
		},
	}
}
