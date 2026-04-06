package cache

import (
	"testing"
	"time"
)

func TestEvaluate(t *testing.T) {
	now := time.Date(2026, 4, 5, 12, 0, 0, 0, time.UTC)
	entry := Entry{
		RefreshBeforeAt: now.Add(30 * time.Second),
		ExpiresAt:       now.Add(60 * time.Second),
	}

	if got := Evaluate(entry, now); got != StateValid {
		t.Fatalf("Evaluate(valid) = %v", got)
	}
	if got := Evaluate(entry, now.Add(45*time.Second)); got != StateRefresh {
		t.Fatalf("Evaluate(refresh) = %v", got)
	}
	if got := Evaluate(entry, now.Add(60*time.Second)); got != StateExpired {
		t.Fatalf("Evaluate(expired) = %v", got)
	}
}

func TestCacheKeyUsesUserAndSkillOnly(t *testing.T) {
	key := Key{
		UserID:  "ou_abc123",
		SkillID: "sales-analysis",
	}

	if got := CacheKey(key); got != "ou_abc123|sales-analysis" {
		t.Fatalf("CacheKey() = %s", got)
	}
}
