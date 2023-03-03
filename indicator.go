package gomimi

type Indicator interface {
	Current() string
	Change(newName string)
}
