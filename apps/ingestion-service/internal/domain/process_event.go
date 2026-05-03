package domain

func BuildIdempotencyKey(tenantID, eventID, eventType string) string {
	return tenantID + ":" + eventID + ":" + eventType
}

func ShouldTriggerReasoning(findings []map[string]interface{}) bool {
	if len(findings) == 0 {
		return false
	}
	for _, finding := range findings {
		if severity, ok := finding["severity"].(string); ok && severity == "high" {
			return true
		}
	}
	return len(findings) > 0
}
