package db

import (
	"github.com/gobwas/glob"
	"strings"
	"sync"
)

var (
	loginDeleteRules   = make([]*LogicDeleteRule, 0)
	loginDeleteRulesMu sync.RWMutex
)

type LogicDeleteRule struct {
	Pattern  string
	Field    string
	SetValue string
	GetValue interface{} // 元素可能为 Cond 或 Union
}

func RegisterLoginDeleteRule(pattern string, deleteRule *LogicDeleteRule) {
	pattern = strings.TrimSpace(pattern)
	if pattern != "" {
		deleteRule.Pattern = pattern
	}
	if deleteRule.Pattern == "" || deleteRule == nil {
		return
	}

	loginDeleteRulesMu.Lock()
	loginDeleteRules = append(loginDeleteRules, deleteRule)
	loginDeleteRulesMu.Unlock()

	matchLogicDeleteRules()
}

func matchLogicDeleteRules(names ...string) {
	metadataMapMu.Lock()
	loginDeleteRulesMu.Lock()
	defer func() {
		metadataMapMu.Unlock()
		loginDeleteRulesMu.Unlock()
	}()

	var nameMap = make(map[string]bool)
	for _, k := range names {
		nameMap[k] = true
	}

	for k, v := range metadataMap {
		if len(names) > 0 && !nameMap[k] {
			continue
		}
		if v.logicDeleteRule != nil && v.logicDeleteRule.Pattern == v.Name {
			// 指定元数据为最高优先级
			continue
		}
		var globalRule *LogicDeleteRule
		for _, rule := range loginDeleteRules {
			if rule.Pattern == "*" {
				globalRule = rule
				break
			}
		}

		for _, rule := range loginDeleteRules {
			if rule.Pattern == "*" {

			}
			g := glob.MustCompile(rule.Pattern)
			if g.Match(v.Name) {
				if v.logicDeleteRule == nil || (v.logicDeleteRule != nil && v.logicDeleteRule.Pattern == "*") {
					v.logicDeleteRule = rule
				}
			} else if globalRule != nil {
				v.logicDeleteRule = globalRule
			}
		}

		metadataMap[k] = v
	}

}
