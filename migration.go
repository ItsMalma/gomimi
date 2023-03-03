package gomimi

type Migration interface {
	Up(builder Builder) error
	Down(builder Builder) error
	Name() string
}
