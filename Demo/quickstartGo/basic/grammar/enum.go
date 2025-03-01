package main

const (
	_, _ Event = iota, iota - 1
	success, duplicateSuccess

	fail, duplicateFail
)
const (
	YES Event = iota + 1
	NO
)

type Event int

func main() {

}
