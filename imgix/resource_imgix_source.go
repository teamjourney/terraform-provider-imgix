package imgix

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"
	"time"
)

func resourceImgixSource() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceSourceRead,
		UpdateContext: resourceSourceUpdate,
		CreateContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return nil
		},
		DeleteContext: func(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
			return nil
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Source display name. Does not impact how images are served.",
			},
			"deployment_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current deployment status. Possible values are deploying, deployed, disabled, and deleted.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether or not a Source is enabled and capable of serving traffic.",
			},
			"date_deployed": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Unix timestamp of when this Source was deployed.",
			},
			"secure_url_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Signing token used for securing images. Only present if deployment.secure_url_enabled is true.",
			},
			"deployment": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allows_upload": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether imgix has the right permissions for this Source to upload to origin.",
						},
						"annotation": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Any comment on the specific deployment.",
						},
						"cache_ttl_behavior": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "respect_origin",
							Description: "Policy to determine how the TTL on imgix images is set.",
							ValidateFunc: validation.StringInSlice([]string{
								"respect_origin",
								"override_origin",
								"enforce_minimum",
							}, false),
						},
						"cache_ttl_error": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      300,
							Description:  "TTL (in seconds) for any error image served when unable to fetch a file from origin.",
							ValidateFunc: validation.IntBetween(1, 31536000),
						},
						"cache_ttl_value": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      31536000,
							Description:  "TTL (in seconds) used by whatever cache mode is set by cache_ttl_behavior.",
							ValidateFunc: validation.IntBetween(1, 31536000),
						},
						"crossdomain_xml_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether this Source should serve a Cross-Domain Policy file if requested.",
						},
						"custom_domains": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: " Non-imgix.net domains you want to use to access your images. Custom domains must be unique across all Sources and must be valid domains.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"default_params": {
							Type:        schema.TypeMap,
							Optional:    true,
							Default:     map[string]string{},
							Description: "Parameters that should be set on all requests to this Source. ",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"image_error": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Image URL imgix should serve instead when a request results in an error.",
						},
						"image_error_append_qs": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether imgix should pass the parameters on the request that received an error to the URL described in image_error",
						},
						"image_missing": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Image URL imgix should serve instead when a request results in a missing image.",
						},
						"image_missing_append_qs": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether imgix should pass the parameters on the request that resulted in a missing image to the URL described in image_missing.",
						},
						"imgix_subdomains": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: "Subdomain you want to use on *.imgix.net to access your images.",
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateSubdomain,
							},
						},
						"secure_url_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether requests must be signed with the secure_url_token to be considered valid.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"azure",
								"gcs",
								"s3",
								"webfolder",
								"webproxy",
							}, false),
						},
						"s3_access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Access Key ID.",
						},
						"s3_secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "S3 Secret Access Key.",
						},
						"s3_bucket": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "S3 bucket name.",
						},
						"s3_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The folder prefix prepended to the image path before resolving the image in S3.",
						},
					},
				},
			},
		},
	}
}

func validateSubdomain(i interface{}, _ cty.Path) diag.Diagnostics {
	domain := i.(string)
	if strings.HasSuffix(domain, "imgix.net") {
		return diag.Errorf("Subdomains can't contain imgix.net suffix. Invalid record: %s", domain)
	}

	return nil
}

func resourceSourceRead(_ context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(*Client)
	var sourceRaw interface{}
	var err error
	if d.IsNewResource() {
		sourceRaw, err = waitForSourceToExist(client, d.Id(), d.Timeout(schema.TimeoutRead))
	} else {
		sourceRaw, _, err = sourceStateRefreshFunc(client, d.Id())()
	}

	if err != nil {
		return diag.Errorf("Error reading source: %s", err.Error())
	}

	source := sourceRaw.(*Source)
	log.Printf("source = %+v", source)

	d.SetId(*source.Id)
	d.Set("name", source.Attributes.Name)
	d.Set("type", source.Type)
	d.Set("deployment_status", source.Attributes.DeploymentStatus)
	d.Set("date_deployed", source.Attributes.DateDeployed)
	d.Set("enabled", source.Attributes.Enabled)
	d.Set("secure_url_token", source.Attributes.SecureUrlToken)
	d.Set("deployment", []interface{}{
		map[string]interface{}{
			"allows_upload":           source.Attributes.Deployment.AllowsUpload,
			"annotation":              source.Attributes.Deployment.Annotation,
			"cache_ttl_behavior":      source.Attributes.Deployment.CacheTtlBehavior,
			"cache_ttl_error":         source.Attributes.Deployment.CacheTtlError,
			"cache_ttl_value":         source.Attributes.Deployment.CacheTtlValue,
			"crossdomain_xml_enabled": source.Attributes.Deployment.CrossdomainXmlEnabled,
			"custom_domains":          source.Attributes.Deployment.CustomDomains,
			"default_params":          source.Attributes.Deployment.DefaultParams,
			"image_error":             source.Attributes.Deployment.ImageError,
			"image_error_append_qs":   source.Attributes.Deployment.ImageErrorAppendQs,
			"image_missing":           source.Attributes.Deployment.ImageMissing,
			"image_missing_append_qs": source.Attributes.Deployment.ImageMissingAppendQs,
			"imgix_subdomains":        source.Attributes.Deployment.ImgixSubdomains,
			"secure_url_enabled":      source.Attributes.Deployment.SecureUrlEnabled,
			"type":                    source.Attributes.Deployment.Type,

			"s3_access_key": source.Attributes.Deployment.S3AccessKey,
			"s3_secret_key": source.Attributes.Deployment.S3SecretKey,
			"s3_bucket":     source.Attributes.Deployment.S3Bucket,
			"s3_prefix":     source.Attributes.Deployment.S3Prefix,
		},
	})

	return nil
}

func resourceSourceUpdate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	source, err := getSourceFromResourceData(d)
	if err != nil {
		return diag.Errorf("Error reading source %s from state: %s", d.Id(), err.Error())
	}

	da, _ := json.MarshalIndent(source, "", "    ")
	log.Print(string(da))

	client := i.(*Client)
	if err := client.updateSource(source); err != nil {
		return diag.Errorf("Error updating source: %s", err)
	}

	return resourceSourceRead(ctx, d, i)
}

func getSourceFromResourceData(d *schema.ResourceData) (*Source, error) {
	deploymentRaw := d.Get("deployment")
	deployments := deploymentRaw.([]interface{})
	if len(deployments) != 1 {
		return nil, errors.New(fmt.Sprintf(
			"Invalid number of deployment elemements in list: %d",
			len(deployments),
		))
	}

	deployment := deployments[0].(map[string]interface{})
	id := d.Id()
	source := &Source{}
	source.Id = &id
	source.Attributes.DateDeployed = Int(d.Get("date_deployed"))
	source.Attributes.DeploymentStatus = String(d.Get("deployment_status"))
	source.Attributes.Enabled = d.Get("enabled").(bool)
	source.Attributes.Name = d.Get("name").(string)
	source.Attributes.SecureUrlToken = String(d.Get("secure_url_token"))
	source.Attributes.Deployment.AllowsUpload = Bool(deployment["allows_upload"])
	source.Attributes.Deployment.Annotation = deployment["annotation"].(string)
	source.Attributes.Deployment.CacheTtlBehavior = deployment["cache_ttl_behavior"].(string)
	source.Attributes.Deployment.CacheTtlError = deployment["cache_ttl_error"].(int)
	source.Attributes.Deployment.CacheTtlValue = deployment["cache_ttl_value"].(int)
	source.Attributes.Deployment.CrossdomainXmlEnabled = deployment["crossdomain_xml_enabled"].(bool)
	source.Attributes.Deployment.CustomDomains = SliceString(deployment["custom_domains"])
	source.Attributes.Deployment.DefaultParams = deployment["default_params"].(map[string]interface{})
	source.Attributes.Deployment.ImageError = String(deployment["image_error"])
	source.Attributes.Deployment.ImageErrorAppendQs = deployment["image_error_append_qs"].(bool)
	source.Attributes.Deployment.ImageMissing = String(deployment["image_missing"])
	source.Attributes.Deployment.ImageMissingAppendQs = deployment["image_missing_append_qs"].(bool)
	source.Attributes.Deployment.ImgixSubdomains = SliceString(deployment["imgix_subdomains"])
	source.Attributes.Deployment.SecureUrlEnabled = Bool(deployment["secure_url_enabled"])
	source.Attributes.Deployment.Type = deployment["type"].(string)
	source.Attributes.Deployment.S3AccessKey = String(deployment["s3_access_key"])
	source.Attributes.Deployment.S3SecretKey = String(deployment["s3_secret_key"])
	source.Attributes.Deployment.S3Bucket = String(deployment["s3_bucket"])
	source.Attributes.Deployment.S3Prefix = String(deployment["s3_prefix"])

	return source, nil
}

func waitForSourceToExist(client *Client, id string, timeout time.Duration) (interface{}, error) {
	log.Printf("[DEBUG] Waiting for source %s being deployed", id)
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deploying"},
		Target:  []string{"deployed"},
		Refresh: sourceStateRefreshFunc(client, id),
		Timeout: timeout,
	}

	return stateConf.WaitForStateContext(context.Background())
}

func sourceStateRefreshFunc(client *Client, id string) resource.StateRefreshFunc {
	return func() (result interface{}, state string, err error) {
		source, err := client.getSourceById(id)
		if err != nil {
			return nil, "", err
		}

		if source == nil {
			return nil, "", errors.New("source not found")
		}

		return source, *source.Attributes.DeploymentStatus, nil
	}
}
