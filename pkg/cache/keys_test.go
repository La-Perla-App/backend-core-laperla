package cache

import "testing"

func TestClusterKeyAndSameSlot(t *testing.T) {
	user := "user-1"
	sess := ClusterKey(user, "sess", "abc")
	idx := ClusterKey(user, "sess", "index")
	if sess != "{user-1}:sess:abc" || idx != "{user-1}:sess:index" {
		t.Fatalf("keys: %s %s", sess, idx)
	}
	if !SameSlot(sess, idx) {
		t.Fatal("expected same slot")
	}
	if SameSlot(sess, ClusterKey("other", "sess", "abc")) {
		t.Fatal("different tags must not share slot")
	}
	if SameSlot("plain-a", "plain-b") {
		t.Fatal("untagged keys are not treated as colocated")
	}
}
