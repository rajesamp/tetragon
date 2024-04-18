// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

// This package contains an end-to-end test for event labels.
package labels_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ec "github.com/cilium/tetragon/api/v1/tetragon/codegen/eventchecker"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"

	sm "github.com/cilium/tetragon/pkg/matchers/stringmatcher"
	"github.com/cilium/tetragon/tests/e2e/checker"
	"github.com/cilium/tetragon/tests/e2e/helpers"
	"github.com/cilium/tetragon/tests/e2e/runners"
)

// This holds our test environment which we get from calling runners.NewRunner().Setup()
var runner *runners.Runner

const (
	// The namespace where we want to spawn our pods
	namespace    = "labels"
	demoAppRetry = 3
)

func installDemoApp(labelsChecker *checker.RPCChecker) features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		manager := helm.New(c.KubeconfigFile())

		for i := 0; i < demoAppRetry; i++ {
			if err := manager.RunUpgrade(
				helm.WithArgs("onlineboutique"),
				helm.WithArgs("oci://us-docker.pkg.dev/online-boutique-ci/charts/onlineboutique"),
				helm.WithArgs("--install"),
				helm.WithArgs("--create-namespace", "-n", namespace),
			); err != nil {
				labelsChecker.ResetTimeout()
				t.Logf("failed to install demo app. run with `-args -v=4` for more context from helm: %s", err)
			} else {
				return ctx
			}
		}

		t.Fatalf("failed to install demo app after %d tries", demoAppRetry)
		return ctx
	}
}

func uninstallDemoApp() features.Func {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		manager := helm.New(c.KubeconfigFile())
		if err := manager.RunUninstall(
			helm.WithName("onlineboutique"),
			helm.WithNamespace(namespace),
		); err != nil {
			t.Fatalf("failed to uninstall demo app. run with `-args -v=4` for more context from helm: %s", err)
		}
		return ctx
	}
}

func TestMain(m *testing.M) {
	runner = runners.NewRunner().Init()

	// Here we ensure our test namespace doesn't already exist then create it.
	runner.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		ctx, _ = helpers.DeleteNamespace(namespace, true)(ctx, c)

		ctx, err := helpers.CreateNamespace(namespace, true)(ctx, c)
		if err != nil {
			return ctx, fmt.Errorf("failed to create namespace: %w", err)
		}

		return ctx, nil
	})

	// Run the tests using the test runner.
	runner.Run(m)
}

func TestLabelsDemoApp(t *testing.T) {
	// Must be called at the beginning of every test
	runner.SetupExport(t)

	labelsChecker := labelsEventChecker().WithEventLimit(5000).WithTimeLimit(5 * time.Minute)

	// This starts labelsChecker and uses it to run event checks.
	runEventChecker := features.New("Run Event Checks").
		Assess("Run Event Checks", labelsChecker.CheckInNamespace(1*time.Minute, namespace)).Feature()

	// This feature waits for labelsChecker to start then runs a custom workload.
	runWorkload := features.New("Run Workload").
		/* Wait up to 30 seconds for the event checker to start before continuing */
		Assess("Wait for Checker", labelsChecker.Wait(30*time.Second)).
		/* Run the workload */
		Assess("Run Workload", installDemoApp(labelsChecker)).
		Feature()

	uninstall := features.New("Uninstall Demo App").
		Assess("Uninstall", uninstallDemoApp()).Feature()

	// Spawn workload and run checker
	runner.TestInParallel(t, runEventChecker, runWorkload)
	runner.Test(t, uninstall)
}

func labelsEventChecker() *checker.RPCChecker {
	labelsEventChecker := ec.NewUnorderedEventChecker(
		ec.NewProcessExecChecker("adservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("adservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("cartservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("cartservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("checkoutservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("checkoutservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("currencyservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("currencyservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("emailservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("emailservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("frontend").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("frontend"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("loadgenerator").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("loadgenerator"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("paymentservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("paymentservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("productcatalogservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("productcatalogservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("recommendationservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("recommendationservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("redis").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("redis-cart"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
		ec.NewProcessExecChecker("shippingservice").WithProcess(ec.NewProcessChecker().WithPod(ec.NewPodChecker().WithPodLabels(map[string]sm.StringMatcher{
			"app":               *sm.Full("shippingservice"),
			"pod-template-hash": *sm.Regex("[a-f0-9]+")}))),
	)

	return checker.NewRPCChecker(labelsEventChecker, "labelsEventChecker")
}
