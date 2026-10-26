package service

import "testing"

func TestExtractCode(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{name: "Chinese sms", input: "您的验证码是123456，请在5分钟内使用。", want: "123456"},
		{name: "Auth keyword", input: "Your auth code is 4567 for login", want: "4567"},
		{name: "OTP", input: "动态密码：654321", want: "654321"},
		{name: "No code", input: "今天的天气不错，明天再说。", want: ""},
		{name: "Short fallback", input: "验证码: 987", want: ""},
	}

	for _, tc := range cases {
		got := ExtractCode(tc.input)
		if got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}

func TestExtractPhoneNumber(t *testing.T) {
	got := ExtractPhoneNumber("SIM1(13800138000)")
	if got != "13800138000" {
		t.Fatalf("got %q want %q", got, "13800138000")
	}
	if got := ExtractPhoneNumber("no number here"); got != "" {
		t.Fatalf("got %q want empty", got)
	}
}
