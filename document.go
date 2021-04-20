package main

type Document interface {
	Text() string
	Category() []string
}
