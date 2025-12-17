package domain

type ConverterFuncTarget struct {
	ID   MyID
	Name string
}

type ConverterFuncTargetPointer struct {
	ID   *MyID
	Name string
}

type ConverterFuncSource struct {
	ID   string
	Name string
}

type ConverterFuncSourcePointer struct {
	ID   *string
	Name string
}

type ConverterFuncTargetSlice struct {
	ID   []MyID
	Name string
}

type ConverterFuncSourceSlice struct {
	ID   []string
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
