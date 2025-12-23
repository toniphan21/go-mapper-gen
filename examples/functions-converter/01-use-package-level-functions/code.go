
package fc

type Target struct {
	ID   MyID
	Name string
}

type Source struct {
	ID   string
	Name string
}

type MyID struct {
	ID string
}

func MyIDToString(v MyID) string {
	return v.ID
}

func StringToMyID(v string) MyID {
	return MyID{ID: v}
}
