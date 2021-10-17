package ex2

import "time"

type Ex1 struct {
	ID        int
	Name      string
	Timestamp time.Time
	Date      time.Time
}

type Ex2 struct {
	Id        int
	Name      string
	Timestamp time.Time
	Date      string
}
