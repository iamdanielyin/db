package db

import (
	"github.com/gobwas/glob"
	"strings"
	"sync"
)

var (
	logicDeleteRuleMap   = make(map[string]*LogicDeleteRule)
	logicDeleteRuleMapMu sync.RWMutex
)

type LogicDeleteRule struct {
	Pattern  string
	Field    string
	SetValue string
	GetValue interface{} // 元素可能为 Cond 或 Union
}

func RegisterLogicDeleteRule(pattern string, rule *LogicDeleteRule) {
	logicDeleteRuleMapMu.Lock()
	defer logicDeleteRuleMapMu.Unlock()

	pattern = strings.TrimSpace(pattern)
	if pattern != "" {
		rule.Pattern = pattern
	}
	if rule.Pattern == "" || rule == nil {
		return
	}

	metadataMapMu.RLock()
	for _, v := range metadataMap {
		g := glob.MustCompile(rule.Pattern)
		if g.Match(v.Name) {
			if exists := logicDeleteRuleMap[v.Name]; exists != nil {
				if exists.Pattern == v.Name {
					// 指定元数据为最高优先级
					continue
				} else if exists.Pattern == "*" && rule.Pattern != "*" {
					logicDeleteRuleMap[v.Name] = rule
				}
			} else {
				logicDeleteRuleMap[v.Name] = rule
			}
			continue
		}
	}
	if _, isMetaRule := metadataMap[rule.Pattern]; !isMetaRule {
		logicDeleteRuleMap[rule.Pattern] = rule
	}
	metadataMapMu.RUnlock()
}

func LookupLogicDeleteRule(name string) *LogicDeleteRule {
	logicDeleteRuleMapMu.RLock()
	defer logicDeleteRuleMapMu.RUnlock()

	return logicDeleteRuleMap[name]
}
