package slice

var internal = struct {
	debugf func(string, ...interface{})
}{
	debugf: func(string, ...interface{}) {
		// do nothing
	},
}
