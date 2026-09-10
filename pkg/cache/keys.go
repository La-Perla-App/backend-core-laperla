package cache

import "strings"

// HashTag wraps a value so Redis Cluster routes related keys to the same slot.
// Only the first `{...}` in a key is the hash tag.
func HashTag(part string) string {
	return "{" + part + "}"
}

// ClusterKey builds `{tag}:a:b`. All keys with the same tag land on one slot,
// so Del/MGet/Eval/SMembers+Del of that family are cluster-safe.
func ClusterKey(tag string, parts ...string) string {
	if len(parts) == 0 {
		return HashTag(tag)
	}
	return HashTag(tag) + ":" + strings.Join(parts, ":")
}

// SameSlot reports whether keys share a hash tag (cheap static check).
// Empty or missing tags are treated as not colocated.
func SameSlot(keys ...string) bool {
	if len(keys) < 2 {
		return true
	}
	tag := hashTagOf(keys[0])
	if tag == "" {
		return false
	}
	for _, k := range keys[1:] {
		if hashTagOf(k) != tag {
			return false
		}
	}
	return true
}

func hashTagOf(key string) string {
	start := strings.IndexByte(key, '{')
	if start < 0 {
		return ""
	}
	end := strings.IndexByte(key[start+1:], '}')
	if end <= 0 {
		return ""
	}
	return key[start+1 : start+1+end]
}
