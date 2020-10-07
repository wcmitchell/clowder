package providers

import (
	"fmt"
	"strings"
	"testing"

	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/config"
	"cloud.redhat.com/clowder/v2/controllers/cloud.redhat.com/errors"
	core "k8s.io/api/core/v1"
)

type TestSecrets struct {
	Secrets      map[string]map[string]string
	ValidSecrets []string
}

func in(candidate string, items []string) bool {
	for _, i := range items {
		if i == candidate {
			return true
		}
	}
	return false
}

func (ts *TestSecrets) ToSecrets(keys []string) []core.Secret {
	secrets := []core.Secret{}

	for k, secMap := range ts.Secrets {
		if !in(k, keys) {
			continue
		}

		bytemap := map[string][]byte{}

		for k, v := range secMap {
			bytemap[k] = []byte(v)
		}

		secrets = append(secrets, core.Secret{
			Data: bytemap,
		})
	}

	return secrets
}

func (ts *TestSecrets) GetExpectedConfig(names []string) *config.ObjectStoreConfig {
	expected := config.ObjectStoreConfig{
		Port:    443,
		Buckets: []config.ObjectStoreBucket{},
	}

	for _, name := range names {
		if ts.IsValid(name) {
			secMap := ts.Secrets[name]
			expected.Buckets = append(expected.Buckets, config.ObjectStoreBucket{
				AccessKey: strPtr(secMap["aws_access_key_id"]),
				SecretKey: strPtr(secMap["aws_secret_access_key"]),
				Name:      secMap["bucket"],
			})

			if val, ok := secMap["endpoint"]; ok {
				expected.Hostname = val
			}
		}
	}

	return &expected
}

func (ts *TestSecrets) IsValid(candidate string) bool {
	for _, i := range ts.ValidSecrets {
		if candidate == i {
			return true
		}
	}

	return false
}

func (ts *TestSecrets) Permutations() []ObjTestParams {
	keys := make([]string, len(ts.Secrets))
	ki := 0

	for k := range ts.Secrets {
		keys[ki] = k
		ki++
	}

	permCnt := exp(2, len(ts.Secrets))
	params := make([]ObjTestParams, permCnt)

	for i := 0; i < permCnt; i++ {
		selectedKeys := []string{}

		for j := 0; j < len(ts.Secrets); j++ {
			if i&(1<<j) > 0 {
				selectedKeys = append(selectedKeys, keys[j])
			}
		}

		params[i] = ObjTestParams{
			Keys:     selectedKeys,
			Secrets:  ts.ToSecrets(selectedKeys),
			Expected: ts.GetExpectedConfig(selectedKeys),
		}
	}

	return params
}

type ObjTestParams struct {
	Keys     []string
	Secrets  []core.Secret
	Expected *config.ObjectStoreConfig
}

func (o *ObjTestParams) ID() string {
	return strings.Join(o.Keys, ":")
}

func (o *ObjTestParams) HasHostname() bool {
	for _, secret := range o.Secrets {
		for k := range secret.Data {
			if k == "endpoint" {
				return true
			}
		}
	}

	return false
}

func exp(base int, exp int) int {
	val := 1

	for i := 0; i < exp; i++ {
		val = val * base
	}

	return val
}

func TestAppInterfaceObjectStore(t *testing.T) {
	testSecretSpecs := TestSecrets{
		Secrets: map[string]map[string]string{
			"ExactKeys": {
				"aws_access_key_id":     "ExactKeys-accessKey",
				"aws_secret_access_key": "ExactKeys-secretKey",
				"aws_region":            "us-east-1",
				"bucket":                "test-bucket",
				"endpoint":              "s3.us-east-1.aws.amazon.com",
			},
			"ExtraKeys": {
				"aws_access_key_id":     "ExtraKeys-accessKey",
				"aws_secret_access_key": "ExtraKeys-secretKey",
				"aws_region":            "us-east-1",
				"bucket":                "extra-bucket",
				"endpoint":              "s3.us-east-1.aws.amazon.com",
				"something":             "else",
			},
			"NoKeys": {},
			"MissingEndpoint": {
				"aws_access_key_id":     "MissingEndpoint-accessKey",
				"aws_secret_access_key": "MissingEndpoint-secretKey",
				"aws_region":            "us-east-1",
				"bucket":                "test-bucket",
			},
			"MissingBucket": {
				"aws_access_key_id":     "MissingBucket-accessKey",
				"aws_secret_access_key": "MissingBucket-secretKey",
				"aws_region":            "us-east-1",
				"endpoint":              "s3.us-east-1.aws.amazon.com",
			},
		},
		ValidSecrets: []string{"ExactKeys", "ExtraKeys", "MissingEndpoint"},
	}

	for _, param := range testSecretSpecs.Permutations() {
		t.Run(param.ID(), func(t *testing.T) {
			c, err := genObjStoreConfig(param.Secrets)

			if param.HasHostname() && err != nil {
				t.Errorf("Error calling genObjStoreConfig: %s", err.(*errors.ClowderError).StackError())
			}

			if len(param.Secrets) > 0 && !param.HasHostname() && err == nil {
				t.Error("genObjStoreConfig should raise an error when hostname is not available")
			}

			if c != nil && len(c.Buckets) > 0 {
				ptr := c.Buckets[0].AccessKey
				if ptr != nil {
					println(fmt.Sprintf("actual: %s", *ptr))
				}
				ptr = param.Expected.Buckets[0].AccessKey
				if ptr != nil {
					println(fmt.Sprintf("expected: %s", *ptr))
				}
			}
			equalsErr := objectStoreEquals(c, param.Expected)

			if equalsErr != "" {
				t.Error(equalsErr)
			}
		})
	}
}

func objectStoreEquals(actual *config.ObjectStoreConfig, expected *config.ObjectStoreConfig) string {
	oneNil, otherNil := actual == nil, expected == nil

	if oneNil && otherNil {
		return ""
	}

	if oneNil != otherNil {
		return "One object is nil"
	}

	actualLen, expectedLen := len(actual.Buckets), len(expected.Buckets)

	if actualLen != expectedLen {
		return fmt.Sprintf("Different number of buckets %d; expected %d", actualLen, expectedLen)
	}

	for i, bucket := range actual.Buckets {
		expectedBucket := expected.Buckets[i]
		if bucket.Name != expectedBucket.Name {
			return fmt.Sprintf("Bad bucket name %s; expected %s", bucket.Name, expectedBucket.Name)
		}
		if *bucket.AccessKey != *expectedBucket.AccessKey {
			return fmt.Sprintf(
				"%s: Bad accessKey '%s'; expected '%s'",
				bucket.Name,
				*bucket.AccessKey,
				*expectedBucket.AccessKey,
			)
		}
		if *bucket.SecretKey != *expectedBucket.SecretKey {
			return fmt.Sprintf(
				"%s: Bad secretKey %s; expected %s",
				bucket.Name,
				*bucket.SecretKey,
				*expectedBucket.SecretKey,
			)
		}
		if bucket.RequestedName != expectedBucket.RequestedName {
			return fmt.Sprintf(
				"%s: Bad requestedName %s; expected %s",
				bucket.Name,
				bucket.RequestedName,
				expectedBucket.RequestedName,
			)
		}
	}

	if actual.Port != expected.Port {
		return fmt.Sprintf("Bad port %d; expected %d", actual.Port, expected.Port)
	}

	if actual.Hostname != expected.Hostname {
		return fmt.Sprintf("Bad hostname %s; expected %s", actual.Hostname, expected.Hostname)
	}

	if actual.AccessKey != expected.AccessKey {
		return fmt.Sprintf("Bad accessKey %s; expected %s", *actual.AccessKey, *expected.AccessKey)
	}

	if actual.SecretKey != expected.SecretKey {
		return fmt.Sprintf("Bad secretKey %s; expected %s", *actual.SecretKey, *expected.SecretKey)
	}

	return ""
}
