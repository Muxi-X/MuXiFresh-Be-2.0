package logic

import (
	"strings"
	"testing"
)

func TestValidateInterviewComment(t *testing.T) {
	if err := validateInterviewComment(""); err != nil {
		t.Fatalf("empty comment should be valid (clears comment), got %v", err)
	}
	if err := validateInterviewComment("正常面评 markdown"); err != nil {
		t.Fatalf("normal comment should be valid, got %v", err)
	}

	// 上限按 rune 计，多字节字符也按 1 个字符算
	atLimit := strings.Repeat("好", maxInterviewCommentLen)
	if err := validateInterviewComment(atLimit); err != nil {
		t.Fatalf("comment at limit should be valid, got %v", err)
	}
	if err := validateInterviewComment(atLimit + "好"); err == nil {
		t.Fatal("comment over limit should be rejected")
	}
}
