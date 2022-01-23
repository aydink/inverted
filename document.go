package inverted

type Document interface {
	Text() string
	Category() []string
}
