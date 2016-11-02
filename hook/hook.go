package hook

import (
	"encoding/json"
	"io/ioutil"
)

type Hooks []Hook

func (h *Hooks) LoadFromFile(path string) error {
	if path == "" {
		return nil
	}

	// parse hook file for hooks
	file, e := ioutil.ReadFile(path)

	if e != nil {
		return e
	}

	e = json.Unmarshal(file, h)
	return e
}
func (h *Hooks) Match(id string) *Hook {
	for i := range *h {
		if (*h)[i].ID == id {
			return &(*h)[i]
		}
	}

	return nil
}

// Hook type is a structure containing details for a single hook
type Hook struct {
	ID                      string `json:"id,omitempty"`
	ExecuteCommand          string `json:"execute-command,omitempty"`
	CommandWorkingDirectory string `json:"command-working-directory,omitempty"`
	ResponseMessage         string `json:"response_message,omitempty"`
	TriggerRule             *Rules `json:"trigger-rule,omitempty"`
}

type MatchRule struct {
	// value, regex
	Type string `json:"type"`

	// tag, branch
	Source string `json:"source"`

	Value string `json:"value"`
}

func (self MatchRule) Evaluate(bbPush *BitPush) (bool, string) {

	compare := self.Value

	source := []string{}
	prefix := ""
	switch self.Source {
	case "tag":
		source = bbPush.GetTags()
		// append the word "tags/" if matching on a tag
		prefix = "tags/"
	case "branch":
		source = bbPush.GetBranches()
	}

	switch self.Type {
	case "value":
		rv, rvs := listHasValue(source, compare)
		return rv, prefix + rvs
	case "regex":
		rv, rvs := listHasRegExValue(source, compare)
		return rv, prefix + rvs
	}

	return false, ""
}

type Rules struct {
	And   *AndRule   `json:"and,omitempty"`
	Or    *OrRule    `json:"or,omitempty"`
	Not   *NotRule   `json:"not,omitempty"`
	Match *MatchRule `json:"match,omitempty"`
}

func (self Rules) Evaluate(bbPush *BitPush) (bool, string) {
	switch {
	case self.And != nil:
		return self.And.Evaluate(bbPush)
	case self.Or != nil:
		return self.Or.Evaluate(bbPush)
	case self.Not != nil:
		return self.Not.Evaluate(bbPush)
	case self.Match != nil:
		return self.Match.Evaluate(bbPush)
	}
	return false, ""
}

type AndRule []Rules

func (self AndRule) Evaluate(bbPush *BitPush) (bool, string) {
	res := true
	resStr := ""

	for _, v := range self {
		rv, rvs := v.Evaluate(bbPush)
		res = res && rv
		resStr = rvs

		if res == false {
			return res, resStr
		}
	}
	return res, resStr
}

type OrRule []Rules

func (self OrRule) Evaluate(bbPush *BitPush) (bool, string) {
	res := false
	resStr := ""

	for _, v := range self {
		rv, rvs := v.Evaluate(bbPush)
		res = res || rv
		resStr = rvs

		if res == true {
			return res, resStr
		}
	}
	return res, resStr
}

type NotRule Rules

func (self NotRule) Evaluate(bbPush *BitPush) (bool, string) {
	rv, rvs := self.Evaluate(bbPush)
	return !rv, rvs
}

func (self *Hook) Evaluate(bbPush *BitPush) (bool, string) {
	switch {
	case self.TriggerRule.And != nil:
		return self.TriggerRule.And.Evaluate(bbPush)
	case self.TriggerRule.Or != nil:
		return self.TriggerRule.Or.Evaluate(bbPush)
	case self.TriggerRule.Not != nil:
		return self.TriggerRule.Not.Evaluate(bbPush)
	case self.TriggerRule.Match != nil:
		return self.TriggerRule.Match.Evaluate(bbPush)
	}

	return false, ""
}
