package detect

import "testing"

func TestKernelCore_UbuntuRelease_KeepsVersionAndAbi(t *testing.T) {
	if got := KernelCore("6.8.0-136-generic"); got != "6.8.0-136" {
		t.Fatalf("expected 6.8.0-136, got %q", got)
	}
}

func TestKernelCore_DebianRelease_DropsTheArchitecture(t *testing.T) {
	if got := KernelCore("6.12.94+deb13-amd64"); got != "6.12.94+deb13" {
		t.Fatalf("expected 6.12.94+deb13, got %q", got)
	}
}

func TestKernelCore_MainlineRelease_IsUnchanged(t *testing.T) {
	if got := KernelCore("6.12.34"); got != "6.12.34" {
		t.Fatalf("expected 6.12.34, got %q", got)
	}
}

func TestKernelCore_Empty_StaysEmpty(t *testing.T) {
	if got := KernelCore(""); got != "" {
		t.Fatalf("expected an empty string, got %q", got)
	}
}

func TestAttachRunningKernel_SetsItOnTheDistributionOnly(t *testing.T) {
	result := &Result{Technologies: []Technology{
		{Name: "Ubuntu", Version: "24.04", Source: SourceOsRelease},
		{Name: "Linux Kernel", Version: "6.8", Source: SourceHostKern},
		{Name: "Laravel", Version: "11"},
	}}

	attachRunningKernel(result, "6.8.0-136")

	if result.Technologies[0].Kernel != "6.8.0-136" {
		t.Fatalf("expected the distribution to carry the build, got %q", result.Technologies[0].Kernel)
	}
	// Upstream point releases and a distribution's ABI counter are different
	// numbers, so the standalone kernel entry must not be given this build.
	if result.Technologies[1].Kernel != "" {
		t.Fatalf("expected the standalone kernel entry to stay empty, got %q", result.Technologies[1].Kernel)
	}
	if result.Technologies[2].Kernel != "" {
		t.Fatalf("expected a non-host technology to stay empty, got %q", result.Technologies[2].Kernel)
	}
}

func TestAttachRunningKernel_UnreadableRelease_ChangesNothing(t *testing.T) {
	result := &Result{Technologies: []Technology{
		{Name: "Ubuntu", Version: "24.04", Kernel: "6.8.0-100", Source: SourceOsRelease},
	}}

	attachRunningKernel(result, "")

	if result.Technologies[0].Kernel != "6.8.0-100" {
		t.Fatalf("expected the existing build to survive, got %q", result.Technologies[0].Kernel)
	}
}
