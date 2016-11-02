package hook

import (
	"log"
	"regexp"
)

// placeholder type to hold values we don't care about
type BitAny map[string]interface{}

// this is the bitbucket webhook message
type BitPush struct {
	//Actor BitAny `json:"actor"`
	//Repo  BitAny `json:"repository"`
	Push BitChanges `json:"push"`
}

func (self *BitPush) GetTags() []string {
	return self.getList("tag")
}

func (self *BitPush) GetBranches() []string {
	return self.getList("branch")
}

func (self *BitPush) getList(t string) []string {
	retVal := []string{}

	for _, change := range self.Push.Changes {
		if change.New.Type == t {
			retVal = append(retVal, change.New.Name)
		}
	}

	return retVal
}

type BitChanges struct {
	Changes []BitChange `json:"changes"`
}

type BitChange struct {
	New     BitRef `json:"new"`
	Old     BitRef `json:"old"`
	Created bool   `json:"created"`
	Closed  bool   `json:"closed"`
}

type BitRef struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func listHasValue(l []string, v string) (bool, string) {
	for _, lv := range l {
		if v == lv {
			return true, lv
		}
	}
	return false, ""
}

func listHasRegExValue(l []string, r string) (bool, string) {
	re, err := regexp.Compile(r)
	if err != nil {
		log.Println("Cannot compile regex: ", r)
		return false, ""
	}

	for _, lv := range l {
		if re.MatchString(lv) {
			return true, lv
		}
	}
	return false, ""
}
