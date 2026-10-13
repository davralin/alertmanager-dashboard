package main

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestSelectedPairsKeepsConfiguredOrder(t *testing.T) {
	got := selectedPairs(map[string]string{
		"job":       "media/rreading-glasses-postgres-v18",
		"namespace": "media",
		"pod":       "rreading-glasses-postgres-v18-1",
		"empty":     "",
	}, []string{"namespace", "pod", "job", "empty", "missing"})

	want := []keyValue{
		{Key: "namespace", Value: "media"},
		{Key: "pod", Value: "rreading-glasses-postgres-v18-1"},
		{Key: "job", Value: "media/rreading-glasses-postgres-v18"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("selectedPairs() = %#v, want %#v", got, want)
	}
}

func TestSortedPairsSortsAndSkipsEmptyValues(t *testing.T) {
	got := sortedPairs(map[string]string{
		"pod":       "rreading-glasses-postgres-v18-1",
		"alertname": "PGReplication",
		"empty":     "",
		"namespace": "media",
	})

	want := []keyValue{
		{Key: "alertname", Value: "PGReplication"},
		{Key: "namespace", Value: "media"},
		{Key: "pod", Value: "rreading-glasses-postgres-v18-1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sortedPairs() = %#v, want %#v", got, want)
	}
}

func TestPageRendersOIDCGreeting(t *testing.T) {
	var output bytes.Buffer
	if err := page.Execute(&output, dashboardPageData{
		DisplayName: "John Doe",
		Username:    "jdoe",
	}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "Hello, John Doe (jdoe)") {
		t.Fatalf("page did not contain OIDC greeting: %s", output.String())
	}
}

func TestPageOmitsGreetingWithoutOIDCIdentity(t *testing.T) {
	var output bytes.Buffer
	if err := page.Execute(&output, dashboardPageData{}); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output.String(), "Hello,") {
		t.Fatalf("page unexpectedly contained OIDC greeting: %s", output.String())
	}
}

func TestPageEscapesOIDCIdentity(t *testing.T) {
	var output bytes.Buffer
	if err := page.Execute(&output, dashboardPageData{
		DisplayName: "<John Doe>",
		Username:    "jdoe&admin",
	}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(output.String(), "Hello, &lt;John Doe&gt; (jdoe&amp;admin)") {
		t.Fatalf("page did not escape OIDC greeting: %s", output.String())
	}
}
