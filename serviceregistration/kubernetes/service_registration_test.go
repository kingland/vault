package kubernetes

import (
	"os"
	"sync"
	"testing"

	"github.com/hashicorp/go-hclog"
	sr "github.com/hashicorp/vault/serviceregistration"
	"github.com/hashicorp/vault/serviceregistration/kubernetes/client"
	kubetest "github.com/hashicorp/vault/serviceregistration/kubernetes/testing"
)

var testVersion = "version 1"

func TestServiceRegistration(t *testing.T) {
	testState, testConf, closeFunc := kubetest.Server(t)
	defer closeFunc()

	client.Scheme = testConf.ClientScheme
	client.TokenFile = testConf.PathToTokenFile
	client.RootCAFile = testConf.PathToRootCAFile
	if err := os.Setenv(client.EnvVarKubernetesServiceHost, testConf.ServiceHost); err != nil {
		t.Fatal(err)
	}
	if err := os.Setenv(client.EnvVarKubernetesServicePort, testConf.ServicePort); err != nil {
		t.Fatal(err)
	}

	if testState.NumPatches() != 0 {
		t.Fatalf("expected 0 patches but have %d: %+v", testState.NumPatches(), testState)
	}
	shutdownCh := make(chan struct{})
	config := map[string]string{
		"namespace": kubetest.ExpectedNamespace,
		"pod_name":  kubetest.ExpectedPodName,
	}
	logger := hclog.NewNullLogger()
	state := sr.State{
		VaultVersion:         testVersion,
		IsInitialized:        true,
		IsSealed:             true,
		IsActive:             true,
		IsPerformanceStandby: true,
	}
	reg, err := NewServiceRegistration(config, logger, state, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := reg.Run(shutdownCh, &sync.WaitGroup{}); err != nil {
		t.Fatal(err)
	}

	// Test initial state.
	if testState.NumPatches() != 5 {
		t.Fatalf("expected 5 current labels but have %d: %+v", testState.NumPatches(), testState)
	}
	if testState.Get(pathToLabels + labelVaultVersion)["value"] != testVersion {
		t.Fatalf("expected %q but received %q", testVersion, testState.Get(pathToLabels + labelVaultVersion)["value"])
	}
	if testState.Get(pathToLabels + labelActive)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelActive)["value"])
	}
	if testState.Get(pathToLabels + labelSealed)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelSealed)["value"])
	}
	if testState.Get(pathToLabels + labelPerfStandby)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelPerfStandby)["value"])
	}
	if testState.Get(pathToLabels + labelInitialized)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelInitialized)["value"])
	}

	// Test NotifyActiveStateChange.
	if err := reg.NotifyActiveStateChange(false); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelActive)["value"] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), testState.Get(pathToLabels + labelActive)["value"])
	}
	if err := reg.NotifyActiveStateChange(true); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelActive)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelActive)["value"])
	}

	// Test NotifySealedStateChange.
	if err := reg.NotifySealedStateChange(false); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelSealed)["value"] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), testState.Get(pathToLabels + labelSealed)["value"])
	}
	if err := reg.NotifySealedStateChange(true); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelSealed)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelSealed)["value"])
	}

	// Test NotifyPerformanceStandbyStateChange.
	if err := reg.NotifyPerformanceStandbyStateChange(false); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelPerfStandby)["value"] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), testState.Get(pathToLabels + labelPerfStandby)["value"])
	}
	if err := reg.NotifyPerformanceStandbyStateChange(true); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelPerfStandby)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelPerfStandby)["value"])
	}

	// Test NotifyInitializedStateChange.
	if err := reg.NotifyInitializedStateChange(false); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelInitialized)["value"] != toString(false) {
		t.Fatalf("expected %q but received %q", toString(false), testState.Get(pathToLabels + labelInitialized)["value"])
	}
	if err := reg.NotifyInitializedStateChange(true); err != nil {
		t.Fatal(err)
	}
	if testState.Get(pathToLabels + labelInitialized)["value"] != toString(true) {
		t.Fatalf("expected %q but received %q", toString(true), testState.Get(pathToLabels + labelInitialized)["value"])
	}
}