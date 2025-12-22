package auth

import (
	"context"
	"log"
)

// Policy：决定“允许哪些认证方式” + 默认优先级
// type Policy struct {
// 	Name            string   // policy/group name
// 	Allowed         []string // allowed strategy names
// 	DefaultStrategy string   // optional
// 	SessionTimeout  int64	 //
//  IdleTimeout		int64	 //
//	RedirectURL		string	 //
// }

type Policy struct {
	Name            string   `yaml:"name" json:"name"`
	Allowed         []string `yaml:"allowed" json:"allowed"`
	DefaultStrategy string   `yaml:"defaultStrategy" json:"default_strategy"`
	SessionTimeout  int      `yaml:"sessionTimeout" json:"session_timeout"`
	IdleTimeout     int      `yaml:"idleTimeout" json:"idle_timeout"`
	RedirectURL     string   `yaml:"redirectURL" json:"redirect_url"`

	// 运行时字段（不序列化）
	Source PolicySource `yaml:"-" json:"-"`
}

type PolicyOverride interface {
	Get(ctx context.Context, username string) (*Policy, error)
}

func logPolicy(user string, p *Policy) {
	log.Printf(
		"[POLICY] user=%s source=%s policy=%s",
		user,
		p.Source,
		p.Name,
	)
}
