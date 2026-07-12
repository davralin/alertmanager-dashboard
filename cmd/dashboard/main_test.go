package main

import (
	"reflect"
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
