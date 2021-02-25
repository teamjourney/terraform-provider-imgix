package imgix

import "testing"

func TestValidatingSubdomains(t *testing.T) {
	cases := map[string]bool{
		"test":             true,
		"test-2":           true,
		"test.imgix.net":   false,
		"test-2.imgix.net": false,
	}

	for c, valid := range cases {
		t.Run(c, func(t *testing.T) {
			res := validateSubdomain(c, nil)
			if res == nil && !valid {
				t.Errorf("Record %s is invalid", c)
			} else if res != nil && valid {
				t.Errorf("Record %s is valid", c)
			}
		})
	}
}
