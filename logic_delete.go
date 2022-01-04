package db

import (
	"github.com/gobwas/glob"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	logicDeleteRuleMap   = make(map[string]*LogicDeleteRule)
	logicDeleteRuleMapMu sync.RWMutex
)

type LogicDeleteRule struct {
	Pattern  string
	SetValue map[string]string
	GetValue Conditional // 元素可能为 Cond 或 Union
}

func (rule *LogicDeleteRule) ParseSetValue() map[string]interface{} {
	doc := make(map[string]interface{})
	for k, v := range rule.SetValue {
		val := rule.parseValue(v)
		if val != nil {
			doc[k] = val
		}
	}
	if len(doc) == 0 {
		return nil
	}
	return doc
}

func (rule *LogicDeleteRule) parseValue(val string) interface{} {
	if val != "" {
		var (
			ss     = len(val)
			vInt   = "$int"
			vFloat = "$float"
			vBool  = "$bool"
		)
		if val == "$now" {
			return time.Now().Unix()
		} else if val == "$now_iso" {
			return time.Now().UTC().Format(time.RFC3339)
		} else if strings.HasPrefix(val, vInt) {
			s := val[len(vInt)+1 : ss-1]
			if v, err := strconv.Atoi(s); err == nil {
				return v
			}
		} else if strings.HasPrefix(val, vFloat) {
			s := val[len(vFloat)+1 : ss-1]
			if v, err := strconv.Atoi(s); err == nil {
				return v
			}
		} else if strings.HasPrefix(val, vBool) {
			s := val[len(vBool)+1 : ss-1]
			if v, err := strconv.ParseBool(s); err == nil {
				return v
			}
		} else {
			return val
		}
	}
	return nil
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
