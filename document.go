package main

type Document interface {
	Id() uint32
	SetId(uint32)
	ParentId() uint32
	Text() string
	Category() []string
}
