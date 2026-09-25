package logic

import (
	"errors"
	"strings"
	"testing"

	"MuXiFresh-Be-2.0/app/form/model"
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

func TestCommentWriteError(t *testing.T) {
	// 文档不存在
	if got := commentWriteError(model.ErrNotFound); got == nil || got.Error() != "entry form not found" {
		t.Fatalf("ErrNotFound should map to entry form not found, got %v", got)
	}
	// CAS 未命中且文档存在 = 版本冲突
	if got := commentWriteError(nil); got == nil || got.Error() != "comment has been modified, please refresh" {
		t.Fatalf("nil findErr should map to conflict, got %v", got)
	}
	// 读回本身出错则原样透传
	sentinel := errors.New("db down")
	if got := commentWriteError(sentinel); !errors.Is(got, sentinel) {
		t.Fatalf("other error should pass through, got %v", got)
	}
}
