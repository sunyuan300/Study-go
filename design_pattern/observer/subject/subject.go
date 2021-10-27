package subject

import "stdlib_learn/design_pattern/observer/observer"

type Subject interface {
	register(o observer.Observer)
	deregister(o observer.Observer)
	notifyAll()
}

type ConCreteSubject struct {

}
