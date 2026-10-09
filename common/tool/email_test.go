package tool

import "testing"

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
