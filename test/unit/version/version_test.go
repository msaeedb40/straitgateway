package version_test

import (
	"strings"
	"testing"

	"github.com/straitgateway/straitgateway/pkg/version"
)

func TestInfo(t *testing.T) {
	info := version.Info()
	if !strings.Contains(info, "StraitGateway") {
		t.Errorf("version.Info() = %q, want to contain 'StraitGateway'", info)
	}
	if !strings.Contains(info, version.Version) {
		t.Errorf("version.Info() = %q, want to contain version %q", info, version.Version)
	}
}

func TestVersionConst(t *testing.T) {
	if version.Version == "" {
		t.Error("version.Version must not be empty")
	}
}
