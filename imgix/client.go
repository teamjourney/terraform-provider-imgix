package imgix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	ApiUrl = "https://api.imgix.com"
)

type Client struct {
	apiKey string
}

type Source struct {
	Id   *string `json:"id,omitempty"`
	Type *string `json:"type,omitempty"`

	Attributes struct {
		DateDeployed     *int    `json:"date_deployed,omitempty"`
		DeploymentStatus *string `json:"deployment_status,omitempty"`
		Enabled          bool    `json:"enabled"`
		Name             string  `json:"name"`
		SecureUrlToken   *string `json:"secure_url_token,omitempty"`

		Deployment struct {
			AllowsUpload          *bool                  `json:"allows_upload,omitempty"`
			Annotation            string                 `json:"annotation"`
			CacheTtlBehavior      string                 `json:"cache_ttl_behavior"`
			CacheTtlError         int                    `json:"cache_ttl_error"`
			CacheTtlValue         int                    `json:"cache_ttl_value"`
			CrossdomainXmlEnabled bool                   `json:"crossdomain_xml_enabled"`
			CustomDomains         []string               `json:"custom_domains"`
			DefaultParams         map[string]interface{} `json:"default_params"`
			ImageError            *string                `json:"image_error"`
			ImageErrorAppendQs    bool                   `json:"image_error_append_qs"`
			ImageMissing          *string                `json:"image_missing"`
			ImageMissingAppendQs  bool                   `json:"image_missing_append_qs"`
			ImgixSubdomains       []string               `json:"imgix_subdomains"`

			S3AccessKey *string `json:"s3_access_key"`
			S3SecretKey *string `json:"s3_secret_key"`
			S3Bucket    *string `json:"s3_bucket"`
			S3Prefix    *string `json:"s3_prefix"`

			SecureUrlEnabled *bool  `json:"secure_url_enabled"`
			Type             string `json:"type"`
		} `json:"deployment"`
	} `json:"attributes"`
}

func (s Source) MarshalJSON() ([]byte, error) {
	type alias Source
	var a = alias(s)
	a.Id = nil
	a.Type = nil
	a.Attributes.DateDeployed = nil
	a.Attributes.DeploymentStatus = nil
	a.Attributes.SecureUrlToken = nil
	a.Attributes.Deployment.AllowsUpload = nil
	return json.Marshal(a)
}

type SourceRequest struct {
	Data *Source `json:"data"`
}

func NewClient(config Config) *Client {
	return &Client{
		apiKey: config.AccessKey,
	}
}

func (c *Client) getSourceById(id string) (*Source, error) {
	req, _ := http.NewRequest(http.MethodGet, ApiUrl+"/api/v1/sources/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	source := &SourceRequest{}

	if err = json.NewDecoder(res.Body).Decode(source); err != nil {
		return nil, err
	}
	return source.Data, nil
}

func (c *Client) updateSource(source *Source) error {
	d := SourceRequest{Data: source}
	b, err := json.Marshal(d)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshalling data: %s", err.Error()))
	}

	req, _ := http.NewRequest(http.MethodPatch, ApiUrl+"/api/v1/sources/"+*source.Id, bytes.NewBuffer(b))
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Error sending request to Imgix API: %s",
			err,
		))
	}

	if res.StatusCode != http.StatusOK {
		return serializeApiError(res)
	}

	return nil
}

func serializeApiError(res *http.Response) error {
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msg := fmt.Sprintf("Error parsing response from request. Status code: %d", res.StatusCode)
		return errors.New(msg)
	}

	return errors.New(fmt.Sprintf(
		"Error response from Imgix API. Status code: %d, error: %s",
		res.StatusCode,
		string(text),
	))
}
