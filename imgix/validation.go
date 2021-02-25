package imgix

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"
)

func validateSubdomain(i interface{}, _ cty.Path) diag.Diagnostics {
	domain := i.(string)
	if strings.HasSuffix(domain, "imgix.net") {
		return diag.Errorf("Subdomain can't contain imgix.net suffix. Invalid record: %s", domain)
	}

	return nil
}
