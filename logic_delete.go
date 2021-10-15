package db

import (
	"sync"
	//"github.com/gobwas/glob"
)

var (
	loginDeleteRuleMap   = make(map[string]*LogicDeleteRule)
	loginDeleteRuleMapMu sync.RWMutex
)

type LogicDeleteRule struct {
	Field    string
	SetValue string
	GetValue interface{} // 元素可能为 Cond 或 Union
}

func RegisterLoginDeleteRule(pattern string, deleteRule *LogicDeleteRule) {
	loginDeleteRuleMapMu.Lock()
	defer loginDeleteRuleMapMu.Unlock()

	loginDeleteRuleMap[pattern] = deleteRule
}

func LookupLoginDeleteRule(name string) *LogicDeleteRule {
	loginDeleteRuleMapMu.RLock()
	defer loginDeleteRuleMapMu.RUnlock()

}
