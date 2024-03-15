package watermob

type Protector interface {
	Protect(fd int) bool
}
