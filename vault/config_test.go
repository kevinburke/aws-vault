package vault_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/99designs/aws-vault/vault"
)

// see http://docs.aws.amazon.com/cli/latest/userguide/cli-multiple-profiles.html
var exampleConfig = []byte(`# an example profile file
[default]
region=us-west-2
output=json

[profile user2]
region=us-east-1
output=text

[profile withsource]
source_profile=user2
region=us-east-1

[profile withmfa]
source_profile=user2
role_arn=arn:aws:iam::4451234513441615400570:role/aws_admin
mfa_serial=arn:aws:iam::1234513441:mfa/blah
region=us-east-1
`)

func newConfigFile(t *testing.T) string {
	f, err := ioutil.TempFile("", "aws-config")
	if err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(f.Name(), exampleConfig, 0600); err != nil {
		t.Fatal(err)
	}
	return f.Name()
}

func TestConfigParsingProfiles(t *testing.T) {
	f := newConfigFile(t)
	defer os.Remove(f)

	cfg, err := vault.LoadConfig(f)
	if err != nil {
		t.Fatal(err)
	}

	var testCases = []struct {
		expected vault.Profile
		ok       bool
	}{
		{vault.Profile{Name: "user2", Region: "us-east-1"}, true},
		{vault.Profile{Name: "withsource", SourceProfile: "user2", Region: "us-east-1"}, true},
		{vault.Profile{
			Name:          "withmfa",
			SourceProfile: "user2",
			Region:        "us-east-1",
			RoleARN:       "arn:aws:iam::4451234513441615400570:role/aws_admin",
			MFASerial:     "arn:aws:iam::1234513441:mfa/blah",
		}, true},
		{vault.Profile{Name: "nopenotthere"}, false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("profile_%s", tc.expected.Name), func(t *testing.T) {
			actual, ok := cfg.Profile(tc.expected.Name)
			if ok != tc.ok {
				t.Fatalf("Expected second param to be %v, got %v", tc.ok, ok)
			}
			if !reflect.DeepEqual(actual, tc.expected) {
				t.Fatalf("Expected %#v, got %#v", tc.expected, actual)
			}
		})
	}
}

func TestConfigParsingDefault(t *testing.T) {
	f := newConfigFile(t)
	defer os.Remove(f)

	cfg, err := vault.LoadConfig(f)
	if err != nil {
		t.Fatal(err)
	}

	def, ok := cfg.Default()
	if !ok {
		t.Fatalf("Expected to find default profile")
	}

	expected := vault.Profile{
		Name:   "default",
		Region: "us-west-2",
	}

	if !reflect.DeepEqual(def, expected) {
		t.Fatalf("Expected %+v, got %+v", expected, def)
	}
}

func TestSourceProfileFromConfig(t *testing.T) {
	f := newConfigFile(t)
	defer os.Remove(f)

	cfg, err := vault.LoadConfig(f)
	if err != nil {
		t.Fatal(err)
	}

	source, ok := cfg.SourceProfile("withmfa")
	if !ok {
		t.Fatalf("Should have found a source")
	}

	if source.Name != "user2" {
		t.Fatalf("Expected source name %q, got %q", "user2", source.Name)
	}
}
