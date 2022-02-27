package imgix

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

const (
	apiUrl = "https://api.imgix.com"

	TypeSource = "sources"

	InvalidAwsAccessKeyError = "aws_access_key"
)

var (
	missingAccessKeyError = errors.New("missing access key")
)

type client struct {
	apiKey string
	apiUrl string
}

type sourceAttributes struct {
	DateDeployed     *int    `json:"date_deployed,omitempty"`
	DeploymentStatus *string `json:"deployment_status,omitempty"`
	Enabled          *bool   `json:"enabled,omitempty"`
	Name             string  `json:"name"`
	SecureUrlToken   *string `json:"secure_url_token,omitempty"`

	Deployment sourceDeployment `json:"deployment"`
}

type sourceDeployment struct {
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

	GCSAccessKey *string `json:"gcs_access_key"`
	GCSSecretKey *string `json:"gcs_secret_key"`
	GCSBucket    *string `json:"gcs_bucket"`
	GCSPrefix    *string `json:"gcs_prefix"`

	SecureUrlEnabled *bool  `json:"secure_url_enabled"`
	Type             string `json:"type"`
}

type Source struct {
	Id   *string `json:"id,omitempty"`
	Type *string `json:"type,omitempty"`

	Attributes sourceAttributes `json:"attributes"`
}

func (s Source) MarshalJSON() ([]byte, error) {
	type alias Source
	var a = alias(s)
	a.Attributes.DateDeployed = nil
	a.Attributes.DeploymentStatus = nil
	a.Attributes.SecureUrlToken = nil
	a.Attributes.Deployment.AllowsUpload = nil
	return json.Marshal(a)
}

type SourceRequest struct {
	Data *Source `json:"data"`
}

func NewClient(config Config) (*client, error) {
	if config.AccessKey == "" {
		return nil, missingAccessKeyError
	}

	if config.ApiBaseUrl == "" {
		config.ApiBaseUrl = apiUrl
	}

	return &client{
		apiKey: config.AccessKey,
		apiUrl: config.ApiBaseUrl,
	}, nil
}

func (c *client) getSourceById(id string) (*Source, error) {
	res, err := c.doRequest("GET", "/api/v1/sources/"+id, nil)
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

func (c *client) createSource(source *Source) (*Source, error) {
	res, err := c.sendSourceRequest("/api/v1/sources", http.MethodPost, source)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusCreated {
		return nil, serializeApiError(res)
	}

	newSource := &Source{}
	_ = json.NewDecoder(res.Body).Decode(newSource)
	return newSource, nil
}

func (c *client) updateSource(source *Source) (*Source, error) {
	res, err := c.sendSourceRequest(
		"/api/v1/sources/"+*source.Id,
		http.MethodPatch,
		source,
	)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, serializeApiError(res)
	}

	return source, nil
}

func (c *client) sendSourceRequest(endpoint, method string, source *Source) (*http.Response, error) {
	d := SourceRequest{Data: source}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error marshalling data: %s", err.Error()))
	}

	res, err := c.doRequest(method, endpoint, bytes.NewBuffer(b))
	if err != nil {
		return res, errors.New(fmt.Sprintf("Error sending request to Imgix API: %s", err))
	}

	return res, nil
}

func (c *client) deleteSource(source *Source) error {
	source.Attributes.Enabled = Bool(false)
	_, err := c.updateSource(source)
	return err
}

func (c *client) doRequest(method, path string, body io.Reader) (*http.Response, error) {
	url := c.apiUrl + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	return http.DefaultClient.Do(req)
}

func serializeApiError(res *http.Response) error {
	text, err := ioutil.ReadAll(res.Body)
	if err != nil {
		msg := fmt.Sprintf("Error parsing response from request. Status code: %d", res.StatusCode)
		return errors.New(msg)
	}

	apiError := &ApiError{}
	if err := json.Unmarshal(text, apiError); err != nil {
		return errors.New("Error parsing response: " + err.Error())
	}

	return apiError
}
