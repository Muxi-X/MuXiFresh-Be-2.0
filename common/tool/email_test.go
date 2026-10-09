package tool

import (
	"errors"
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"域名大写转小写", "Xxxx@QQ.COM", "Xxxx@qq.com"},
		{"本地部分大小写保留", "AbC@GmAil.COM", "AbC@gmail.com"},
		{"域名混合大小写", "user@Example.Co.Uk", "user@example.co.uk"},
		{"去首尾空白", "  a@b.com  ", "a@b.com"},
		{"已小写不变", "a@b.com", "a@b.com"},
		{"无@仅去空白", "  not-an-email  ", "not-an-email"},
		{"空串", "", ""},
		{"多个@取最后一个为界", "a@b@QQ.COM", "a@b@qq.com"},
		{"域名已小写保持", "user@qq.com", "user@qq.com"},
		{"@后为空", "user@", "user@"},
	}
	for _, c := range cases {
		if got := NormalizeEmail(c.in); got != c.want {
			t.Errorf("%s: NormalizeEmail(%q)=%q, want %q", c.name, c.in, got, c.want)
		}
	}
}

func TestValidateEmail(t *testing.T) {
	ok := []struct {
		name string
		in   string
		want string
	}{
		{"合法", "a@b.com", "a@b.com"},
		{"域名大写转小写", "User@QQ.COM", "User@qq.com"},
		{"去首尾空白", "  a@b.com ", "a@b.com"},
		{"多点域名", "user@mail.co.uk", "user@mail.co.uk"},
		{"连字符域名", "user@my-site.com", "user@my-site.com"},
	}
	for _, c := range ok {
		got, err := ValidateEmail(c.in)
		if err != nil {
			t.Errorf("%s: unexpected error %v", c.name, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: ValidateEmail(%q)=%q, want %q", c.name, c.in, got, c.want)
		}
	}

	bad := []struct {
		name string
		in   string
	}{
		{"空串", ""},
		{"纯空白", "   "},
		{"无@", "a"},
		{"无点域名", "a@b"},
		{"带显示名", "Name <a@b.com>"},
		{"@后为空", "a@"},
		{"域名含非法字符", "a@bad!.com"},
		{"域名label以连字符开头", "a@-bad.com"},
		{"域名label以连字符结尾", "a@bad-.com"},
	}
	for _, c := range bad {
		if _, err := ValidateEmail(c.in); !errors.Is(err, ErrInvalidEmail) {
			t.Errorf("%s: expected ErrInvalidEmail, got %v", c.name, err)
		}
	}
}
