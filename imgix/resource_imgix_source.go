package imgix

import (
	"context"
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
		Description:   "Allows managing Imgix sources",
		ReadContext:   resourceSourceRead,
		UpdateContext: resourceSourceUpdate,
		CreateContext: resourceSourceCreate,
		DeleteContext: resourceSourceDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Minute * 30),
			Update: schema.DefaultTimeout(time.Minute * 30),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
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
			"wait_for_deployed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Determines if Terraform should wait for deployed status after any change",
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
							Sensitive:   true,
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

	if d.Get("wait_for_deployed").(bool) {
		sourceRaw, err = waitForSourceToBeDeployed(client, d.Id(), d.Timeout(schema.TimeoutRead))
	} else {
		sourceRaw, _, err = sourceStateRefreshFunc(client, d.Id())()
	}

	if err != nil {
		return diag.Errorf("Error reading source: %s", err.Error())
	}

	source := sourceRaw.(*Source)

	d.SetId(*source.Id)
	d.Set("name", source.Attributes.Name)
	d.Set("type", source.Type)
	d.Set("deployment_status", source.Attributes.DeploymentStatus)
	d.Set("date_deployed", source.Attributes.DateDeployed)
	d.Set("enabled", source.Attributes.Enabled)
	d.Set("secure_url_token", source.Attributes.SecureUrlToken)
	deployment := map[string]interface{}{}
	if deploymentRaw, ok := d.GetOk("deployment"); ok {
		if deploymentRaw != nil {
			deployment = deploymentRaw.([]interface{})[0].(map[string]interface{})
		}
	}

	deployment["allows_upload"] = source.Attributes.Deployment.AllowsUpload
	deployment["annotation"] = source.Attributes.Deployment.Annotation
	deployment["cache_ttl_behavior"] = source.Attributes.Deployment.CacheTtlBehavior
	deployment["cache_ttl_error"] = source.Attributes.Deployment.CacheTtlError
	deployment["cache_ttl_value"] = source.Attributes.Deployment.CacheTtlValue
	deployment["crossdomain_xml_enabled"] = source.Attributes.Deployment.CrossdomainXmlEnabled
	deployment["custom_domains"] = source.Attributes.Deployment.CustomDomains
	deployment["default_params"] = source.Attributes.Deployment.DefaultParams
	deployment["image_error"] = source.Attributes.Deployment.ImageError
	deployment["image_error_append_qs"] = source.Attributes.Deployment.ImageErrorAppendQs
	deployment["image_missing"] = source.Attributes.Deployment.ImageMissing
	deployment["image_missing_append_qs"] = source.Attributes.Deployment.ImageMissingAppendQs
	deployment["imgix_subdomains"] = source.Attributes.Deployment.ImgixSubdomains
	deployment["secure_url_enabled"] = source.Attributes.Deployment.SecureUrlEnabled
	deployment["type"] = source.Attributes.Deployment.Type
	deployment["s3_access_key"] = source.Attributes.Deployment.S3AccessKey
	deployment["s3_bucket"] = source.Attributes.Deployment.S3Bucket
	deployment["s3_prefix"] = source.Attributes.Deployment.S3Prefix

	d.Set("deployment", []interface{}{deployment})

	return nil
}

func resourceSourceUpdate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	source, err := getSourceFromResourceData(d)
	if err != nil {
		return diag.Errorf("Error reading source %s from state: %s", d.Id(), err.Error())
	}

	client := i.(*Client)
	source, err = makeSourceRequest(ctx, func() (*Source, error) {
		return client.updateSource(source)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSourceRead(ctx, d, i)
}

func resourceSourceCreate(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	source, err := getSourceFromResourceData(d)
	if err != nil {
		return diag.Errorf("Error reading source %s from state: %s", d.Id(), err.Error())
	}

	source.Id = nil
	source.Attributes.Enabled = nil
	source.Type = String(TypeSource)

	client := i.(*Client)
	newSource, err := makeSourceRequest(ctx, func() (*Source, error) {
		return client.createSource(source)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*newSource.Id)

	return resourceSourceRead(ctx, d, i)
}

func resourceSourceDelete(_ context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(*Client)
	source, err := getSourceFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.deleteSource(source); err != nil {
		return diag.FromErr(err)
	}

	return diag.Diagnostics{
		{
			Severity: diag.Warning,
			Summary:  "Source was disabled because cannot be removed",
			Detail:   "Contact support for actual deletion of sources",
		},
	}
}

func makeSourceRequest(ctx context.Context, operation func() (*Source, error)) (*Source, error) {
	conf := &resource.StateChangeConf{
		Pending: []string{"retry"},
		Target:  []string{"ok"},
		Delay:   time.Second * 3,
		Timeout: time.Second * 10,
		Refresh: func() (interface{}, string, error) {
			source, err := operation()
			if err != nil {
				if isImgixApiErr(err, InvalidAwsAccessKeyError) {
					return nil, "retry", err
				}

				return nil, "error", err
			}

			return source, "ok", nil
		},
	}

	res, err := conf.WaitForStateContext(ctx)
	var s *Source
	if res != nil {
		s = res.(*Source)
	}
	return s, err
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
	source.Type = String(d.Get("type"))
	source.Attributes.DateDeployed = Int(d.Get("date_deployed"))
	source.Attributes.DeploymentStatus = String(d.Get("deployment_status"))
	source.Attributes.Enabled = Bool(d.Get("enabled"))
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
	source.Attributes.Deployment.ImageError = StringNilIfEmpty(deployment["image_error"])
	source.Attributes.Deployment.ImageErrorAppendQs = deployment["image_error_append_qs"].(bool)
	source.Attributes.Deployment.ImageMissing = StringNilIfEmpty(deployment["image_missing"])
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

func waitForSourceToBeDeployed(client *Client, id string, timeout time.Duration) (*Source, error) {
	log.Printf("[DEBUG] Waiting for source %s being deployed", id)
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deploying"},
		Target:  []string{"deployed"},
		// source doesn't start deploying immediately after request is finished
		Delay:   5 * time.Second,
		Refresh: sourceStateRefreshFunc(client, id),
		Timeout: timeout,
	}

	res, err := stateConf.WaitForStateContext(context.Background())
	var source *Source
	if res != nil {
		source = res.(*Source)
	}
	return source, err
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

		log.Printf(
			"[TRACE] Source %s deployment status: %s",
			*source.Id,
			*source.Attributes.DeploymentStatus,
		)

		return source, *source.Attributes.DeploymentStatus, nil
	}
}
