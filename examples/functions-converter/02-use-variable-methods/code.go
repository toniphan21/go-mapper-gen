
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

type convertFuncs struct{}

func (c *convertFuncs) MyIDToString(v MyID) string {
	return v.ID
}

func (c *convertFuncs) StringToMyID(v string) MyID {
	return MyID{ID: v}
}

var ConvertFuncs = &convertFuncs{}
