package main

import "testing"

func TestGroupKeyTracksBusinessAction(t *testing.T) {
	cases := []struct {
		name   string
		event  MatterEvent
		wanted string
	}{
		{"intake", MatterEvent{MatterID: "m-17", Action: "matter_intake"}, "matter_intake"},
		{"delivery", MatterEvent{MatterID: "m-17", Action: "signed_document_delivery"}, "signed_document_delivery"},
		{"deadline", MatterEvent{MatterID: "m-17", Action: "deadline_follow_up"}, "deadline_follow_up"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := groupKey(tc.event); got != tc.wanted {
				t.Fatalf("group key = %q, want %q", got, tc.wanted)
			}
		})
	}
}
