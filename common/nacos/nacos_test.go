package nacos

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func TestServiceUsesConfiguredGroup(t *testing.T) {
	t.Setenv("NACOS_GROUP", "TEST")

	target := struct{}{}
	option := Service("user-api", &target)

	if option.Group != "TEST" {
		t.Fatalf("expected group TEST, got %q", option.Group)
	}
	if option.DataId != "user-api" {
		t.Fatalf("expected data ID user-api, got %q", option.DataId)
	}
}

func TestInfraCanUseIndependentLocation(t *testing.T) {
	t.Setenv("NACOS_GROUP", "TEST")
	t.Setenv("NACOS_INFRA_GROUP", "INFRA")
	t.Setenv("NACOS_INFRA_DATA_ID", "infra-test")

	target := struct{}{}
	option := Infra(&target)

	if option.Group != "INFRA" {
		t.Fatalf("expected group INFRA, got %q", option.Group)
	}
	if option.DataId != "infra-test" {
		t.Fatalf("expected data ID infra-test, got %q", option.DataId)
	}
}

func TestInfraDefaultsToServiceGroup(t *testing.T) {
	t.Setenv("NACOS_GROUP", "DEV")
	t.Setenv("NACOS_INFRA_GROUP", "")
	t.Setenv("NACOS_INFRA_DATA_ID", "")

	option := Infra(&struct{}{})

	if option.Group != "DEV" {
		t.Fatalf("expected group DEV, got %q", option.Group)
	}
	if option.DataId != defaultInfraDataID {
		t.Fatalf("expected data ID %q, got %q", defaultInfraDataID, option.DataId)
	}
}

func TestNacosContentUsesYAML(t *testing.T) {
	var target struct {
		MongoDB struct {
			URL string
			DB  string
		}
	}

	content := []byte("MongoDB:\n  URL: mongodb://mongodb:27017\n  DB: muxi_fresh\n")
	if err := yaml.Unmarshal(content, &target); err != nil {
		t.Fatalf("load YAML: %v", err)
	}

	if target.MongoDB.URL != "mongodb://mongodb:27017" {
		t.Fatalf("unexpected MongoDB URL %q", target.MongoDB.URL)
	}
	if target.MongoDB.DB != "muxi_fresh" {
		t.Fatalf("unexpected MongoDB database %q", target.MongoDB.DB)
	}
}

func TestYAMLAllowsServiceOwnedEtcdKeyWithoutSharedConnection(t *testing.T) {
	var target struct {
		Infra struct {
			Etcd struct {
				Hosts []string
			}
		}
		UserConf struct {
			Etcd struct {
				Hosts []string
				Key   string
			}
		}
	}

	content := []byte("UserConf:\n  Etcd:\n    Key: user.rpc\n")
	if err := yaml.Unmarshal(content, &target); err != nil {
		t.Fatalf("load partial service YAML: %v", err)
	}

	if target.UserConf.Etcd.Key != "user.rpc" {
		t.Fatalf("unexpected Etcd key %q", target.UserConf.Etcd.Key)
	}
	if len(target.UserConf.Etcd.Hosts) != 0 {
		t.Fatalf("service YAML unexpectedly populated Etcd hosts: %v", target.UserConf.Etcd.Hosts)
	}
}
