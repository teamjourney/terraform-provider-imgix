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
			client := i.(*Client)
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
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
		},
	}
}
