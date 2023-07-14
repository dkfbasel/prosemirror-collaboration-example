package prosemirror

// Permission indicates the users permissions on a specific document
type Permission int

// initialize available permissions
const (
	None Permission = iota
	Comment
	Edit
)

func (p Permission) String() string {
	return [...]string{"none", "comment", "edit"}[p]
}
