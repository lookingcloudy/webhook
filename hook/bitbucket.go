package hook

type BitAny map[string]interface{}

type BitPush struct {
	//Actor BitAny `json:"actor"`
	//Repo  BitAny `json:"repository"`
	Push BitChanges `json:"push"`

}

type BitChanges struct {
	Changes []BitChange `json:"changes"`
}

type BitChange struct {
	New BitRef `json:"new"`
	Old BitRef `json:"old"`
	Created bool `json:"created"`
	Closed bool `json:"closed"`
}

type BitRef struct {
	Type string `json:"type"`
	Name string `json:"name"`
}
