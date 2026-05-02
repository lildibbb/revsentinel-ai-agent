package domain

import "testing"

func TestBuildIdempotencyKey_StableFormat(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		tenantID  string
		eventID   string
		eventType string
		want      string
	}{
		{
			name:      "normal values",
			tenantID:  "demo-tenant",
			eventID:   "evt-1",
			eventType: "discount_event",
			want:      "demo-tenant:evt-1:discount_event",
		},
		{
			name:      "empty values still deterministic",
			tenantID:  "",
			eventID:   "",
			eventType: "",
			want:      "::",
		},
		{
			name:      "whitespace preserved",
			tenantID:  "tenant a",
			eventID:   "event 2",
			eventType: "type b",
			want:      "tenant a:event 2:type b",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := BuildIdempotencyKey(tc.tenantID, tc.eventID, tc.eventType)
			if got != tc.want {
				t.Fatalf("want %q got %q", tc.want, got)
			}
		})
	}
}
