package generator

type Author struct {
	Name        string
	ContactInfo []AuthorContact
}

type AuthorContact struct {
	Medium string
	Value  string
}
