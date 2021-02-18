package imgix

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	testSourceId       = "601430223753592c4e822e2c"
	testSourceEndpoint = "/api/v1/sources/" + testSourceId

	testApiToken = "abc"
)

func TestCreatingClientWithoutKey(t *testing.T) {
	c, e := NewClient(Config{
		AccessKey: "",
	})

	if e == nil {
		t.Error("error should not be nil")
	}

	if e != missingAccessKeyError {
		t.Error("invalid error when access key is empty")
	}

	if c != nil {
		t.Error("client should be empty")
	}
}

func TestCreatingClientWithKey(t *testing.T) {
	key := testApiToken
	c, e := NewClient(Config{
		AccessKey: key,
	})

	if e != nil {
		t.Error("error should be nil")
		return
	}

	if c == nil {
		t.Error("client should not be empty")
	}

	if c.apiKey != key {
		t.Error("api key does not match")
	}
}

func startMockServer(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		path := req.URL.Path

		if a := req.Header.Get("Authorization"); a != "Bearer "+testApiToken {
			t.Error("invalid authorization header")
			return
		}

		if path == testSourceEndpoint {
			rawJson, err := ioutil.ReadFile("./testdata/sample_source.json")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(rawJson)
		} else if path == testSourceEndpoint {

		}
	}))

	return ts
}

func prepareHttpTest(t *testing.T) *client {
	ts := startMockServer(t)
	t.Cleanup(func() {
		ts.Close()
	})

	c, e := NewClient(Config{
		AccessKey:  testApiToken,
		ApiBaseUrl: ts.URL,
	})

	if e != nil {
		t.Error("creating client error should be nil")
	}

	return c
}

func TestGettingSourceById(t *testing.T) {
	c := prepareHttpTest(t)
	s, err := c.getSourceById(testSourceId)
	if err != nil {
		t.Error("response error should be nil")
		return
	}

	expected := &Source{
		Id:   String(testSourceId),
		Type: String(TypeSource),
		Attributes: sourceAttributes{
			DateDeployed:     Int(1612274615),
			DeploymentStatus: String("disabled"),
			Enabled:          Bool(false),
			Name:             "source1",
			Deployment: sourceDeployment{
				Annotation:            "source1 annotation",
				CacheTtlBehavior:      "respect_origin",
				CacheTtlError:         300,
				CacheTtlValue:         31536000,
				CrossdomainXmlEnabled: false,
				CustomDomains:         []string{},
				DefaultParams:         map[string]interface{}{},
				ImageError:            nil,
				ImageErrorAppendQs:    false,
				ImageMissing:          nil,
				ImageMissingAppendQs:  false,
				ImgixSubdomains:       []string{"example-1", "example-2"},
				S3AccessKey:           String("AKIABCDEFGHI"),
				S3SecretKey:           nil,
				S3Bucket:              String("abc-bucket"),
				S3Prefix:              String("imgix-files"),
				SecureUrlEnabled:      Bool(false),
				Type:                  "s3",
			},
		},
	}

	if !reflect.DeepEqual(s, expected) {
		t.Error("source doesnt match expected")
	}
}

func TestDeletingSource(t *testing.T) {
	c := prepareHttpTest(t)

	source := &Source{
		Id: String(testSourceId),
		Attributes: sourceAttributes{
			Enabled: Bool(true),
		},
	}

	e := c.deleteSource(source)
	if e != nil {
		t.Error("error should be nil when deleting source")
	}

	if *source.Attributes.Enabled == true {
		t.Error("source should be disabled after deletion")
	}
}
