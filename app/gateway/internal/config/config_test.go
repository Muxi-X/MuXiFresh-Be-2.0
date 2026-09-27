package config

import "testing"

func TestApplyInfraPinsTransparencyMiddlewares(t *testing.T) {
	c := Config{}
	c.Infra.Middlewares.Timeout = true
	c.Infra.Middlewares.Gunzip = true
	c.Infra.Middlewares.Recover = true

	c.ApplyInfra()

	if c.RestConf.Middlewares.Timeout {
		t.Fatal("Timeout middleware must stay disabled for transparent proxying")
	}
	if c.RestConf.Middlewares.Gunzip {
		t.Fatal("Gunzip middleware must stay disabled for transparent proxying")
	}
	if !c.RestConf.Middlewares.Recover {
		t.Fatal("Recover middleware should be inherited from infra")
	}
}

func TestApplyInfraEnablesMaxBytesOnlyWithLimit(t *testing.T) {
	withLimit := Config{}
	withLimit.MaxBytes = 8 << 20
	withLimit.ApplyInfra()
	if !withLimit.Middlewares.MaxBytes {
		t.Fatal("MaxBytes middleware should be enabled when a positive limit is set")
	}

	withoutLimit := Config{}
	withoutLimit.Infra.Middlewares.MaxBytes = true
	withoutLimit.ApplyInfra()
	if withoutLimit.Middlewares.MaxBytes {
		t.Fatal("MaxBytes middleware should stay disabled when no positive limit is set")
	}
}
