package schema

type Author struct {
	Name        string          `json:"name"`
	ContactInfo []AuthorContact `json:"contactInfo,omitempty"`
}

type AuthorContact struct {
	Medium string `json:"medium"`
	Value  string `json:"value"`
}
