package auth

import (
	"testing"
	"time"
)

func TestTokenIssueAndParse(t *testing.T) {
	svc, err := NewTokenService("0123456789abcdef", "go-starter", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	token, err := svc.Issue(42, []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}

	claims, err := svc.Parse(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != 42 || len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Fatalf("claims 不符: %+v", claims)
	}
}

func TestTokenExpired(t *testing.T) {
	if _, err := NewTokenService("0123456789abcdef", "go-starter", -time.Minute); err == nil {
		t.Fatal("ttl 为负应当初始化失败")
	}

	valid, err := NewTokenService("0123456789abcdef", "go-starter", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := valid.Issue(1, nil)
	if err != nil {
		t.Fatal(err)
	}

	// 用另一个 issuer 校验，模拟 token 被伪造/跨环境复用
	other, err := NewTokenService("0123456789abcdef", "another-issuer", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.Parse(token); err == nil {
		t.Fatal("issuer 不匹配时应校验失败")
	}
}
