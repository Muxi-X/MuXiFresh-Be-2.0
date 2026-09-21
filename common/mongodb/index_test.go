package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestKeysMatch(t *testing.T) {
	want := bson.D{{Key: "user_id", Value: 1}}

	cases := []struct {
		name string
		got  bson.D
		want bool
	}{
		{"驱动解码为 int32 时匹配", bson.D{{Key: "user_id", Value: int32(1)}}, true},
		{"驱动解码为 int64 时匹配", bson.D{{Key: "user_id", Value: int64(1)}}, true},
		{"驱动解码为 int 时匹配", bson.D{{Key: "user_id", Value: 1}}, true},
		{"降序索引不匹配", bson.D{{Key: "user_id", Value: int32(-1)}}, false},
		{"键名不同不匹配", bson.D{{Key: "other", Value: int32(1)}}, false},
		{"键数量不同不匹配", bson.D{{Key: "user_id", Value: int32(1)}, {Key: "cycle", Value: int32(1)}}, false},
	}
	for _, c := range cases {
		if got := keysMatch(c.got, want); got != c.want {
			t.Errorf("%s: keysMatch=%v, want %v", c.name, got, c.want)
		}
	}
}

func TestKeysMatch_OrderSensitive(t *testing.T) {
	// 复合索引的键顺序是语义的一部分：交换顺序应视为不匹配
	want := bson.D{{Key: "user_id", Value: 1}, {Key: "cycle", Value: 1}}
	swapped := bson.D{{Key: "cycle", Value: 1}, {Key: "user_id", Value: 1}}

	if !keysMatch(want, want) {
		t.Error("identical compound keys should match")
	}
	if keysMatch(swapped, want) {
		t.Error("swapped key order must not match")
	}
	if keysMatch(bson.D{{Key: "user_id", Value: 1}, {Key: "cycle", Value: -1}}, want) {
		t.Error("differing sort order should not match")
	}
}

func TestIndexMatches(t *testing.T) {
	spec := IndexSpec{
		Collection: "entry_form",
		Name:       "entry_form_user_id_cycle",
		Unique:     true,
		Keys:       bson.D{{Key: "user_id", Value: 1}, {Key: "cycle", Value: 1}},
	}

	same := &indexInfo{Name: spec.Name, Unique: true, Key: bson.D{{Key: "user_id", Value: int32(1)}, {Key: "cycle", Value: int32(1)}}}
	if !indexMatches(same, spec) {
		t.Error("identical definition should match (int32 vs int normalized)")
	}

	notUnique := &indexInfo{Name: spec.Name, Unique: false, Key: bson.D{{Key: "user_id", Value: 1}, {Key: "cycle", Value: 1}}}
	if indexMatches(notUnique, spec) {
		t.Error("non-unique index must not match a unique spec")
	}

	otherKeys := &indexInfo{Name: spec.Name, Unique: true, Key: bson.D{{Key: "user_id", Value: 1}}}
	if indexMatches(otherKeys, spec) {
		t.Error("different keys must not match")
	}
}
