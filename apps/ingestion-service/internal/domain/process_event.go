package domain

func BuildIdempotencyKey(tenantID, eventID, eventType string) string {
	return tenantID + ":" + eventID + ":" + eventType
}
