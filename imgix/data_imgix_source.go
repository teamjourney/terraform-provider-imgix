package imgix

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceImgixSource() *schema.Resource {
	return &schema.Resource{
		Description: "Allows getting Imgix source information",
		ReadContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			client := i.(*client)
			id := data.Get("id").(string)

			source, err := client.getSourceById(id)
			if err != nil {
				return diag.FromErr(err)
			}

			data.Set("name", source.Attributes.Name)
			data.SetId(id)

			return nil
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: sourceDescriptions["id"],
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["type"],
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["name"],
			},
			"deployment_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["deployment_status"],
			},
			"date_deployed": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: sourceDescriptions["date_deployed"],
			},
			"deployment": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allows_upload": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: sourceDescriptions["allows_upload"],
						},
						"annotation": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["annotation"],
						},
						"cache_ttl_behavior": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["cache_ttl_behavior"],
						},
						"cache_ttl_error": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: sourceDescriptions["cache_ttl_error"],
						},
						"cache_ttl_value": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: sourceDescriptions["cache_ttl_value"],
						},
						"crossdomain_xml_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: sourceDescriptions["crossdomain_xml_enabled"],
						},
						"custom_domains": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: sourceDescriptions["custom_domains"],
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"default_params": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: sourceDescriptions["default_params"],
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"image_error": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["image_error"],
						},
						"image_error_append_qs": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: sourceDescriptions["image_error_append_qs"],
						},
						"image_missing": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["image_missing"],
						},
						"image_missing_append_qs": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: sourceDescriptions["image_missing_append_qs"],
						},
						"imgix_subdomains": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: sourceDescriptions["imgix_subdomains"],
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"secure_url_enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: sourceDescriptions["secure_url_enabled"],
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["deployment_type"],
						},
						"s3_access_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["s3_access_key"],
						},
						"s3_secret_key": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["s3_secret_key"],
						},
						"s3_bucket": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["s3_bucket"],
						},
						"s3_prefix": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: sourceDescriptions["s3_prefix"],
						},
					},
				},
			},
		},
	}
}
