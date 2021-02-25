package imgix

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["id"],
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["type"],
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: sourceDescriptions["name"],
			},
			"deployment_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["deployment_status"],
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: sourceDescriptions["enabled"],
			},
			"date_deployed": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: sourceDescriptions["date_deployed"],
			},
			"secure_url_token": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: sourceDescriptions["secure_url_token"],
			},
			"wait_for_deployed": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: sourceDescriptions["wait_for_deployed"],
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
							Description: sourceDescriptions["allows_upload"],
						},
						"annotation": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["annotation"],
						},
						"cache_ttl_behavior": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "respect_origin",
							Description: sourceDescriptions["cache_ttl_behavior"],
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
							Description:  sourceDescriptions["cache_ttl_error"],
							ValidateFunc: validation.IntBetween(1, 31536000),
						},
						"cache_ttl_value": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      31536000,
							Description:  sourceDescriptions["cache_ttl_value"],
							ValidateFunc: validation.IntBetween(1, 31536000),
						},
						"crossdomain_xml_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: sourceDescriptions["crossdomain_xml_enabled"],
						},
						"custom_domains": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: sourceDescriptions["custom_domains"],
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"default_params": {
							Type:        schema.TypeMap,
							Optional:    true,
							Default:     map[string]string{},
							Description: sourceDescriptions["default_params"],
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"image_error": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["image_error"],
						},
						"image_error_append_qs": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: sourceDescriptions["image_error_append_qs"],
						},
						"image_missing": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["image_missing"],
						},
						"image_missing_append_qs": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: sourceDescriptions["image_missing_append_qs"],
						},
						"imgix_subdomains": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							Description: sourceDescriptions["imgix_subdomains"],
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validateSubdomain,
							},
						},
						"secure_url_enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: sourceDescriptions["secure_url_enabled"],
						},
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: sourceDescriptions["deployment_type"],
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
							Description: sourceDescriptions["s3_access_key"],
						},
						"s3_secret_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["s3_secret_key"],
							Sensitive:   true,
						},
						"s3_bucket": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["s3_bucket"],
						},
						"s3_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: sourceDescriptions["s3_prefix"],
						},
					},
				},
			},
		},
	}
}

func resourceSourceRead(_ context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	client := i.(*client)
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

	c := i.(*client)
	source, err = makeSourceRequest(ctx, func() (*Source, error) {
		return c.updateSource(source)
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

	c := i.(*client)
	newSource, err := makeSourceRequest(ctx, func() (*Source, error) {
		return c.createSource(source)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(*newSource.Id)

	return resourceSourceRead(ctx, d, i)
}

func resourceSourceDelete(_ context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
	c := i.(*client)
	source, err := getSourceFromResourceData(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if delErr := c.deleteSource(source); delErr != nil {
		return diag.FromErr(delErr)
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
				return nil, "", err
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

func waitForSourceToBeDeployed(client *client, id string, timeout time.Duration) (*Source, error) {
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

func sourceStateRefreshFunc(client *client, id string) resource.StateRefreshFunc {
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
