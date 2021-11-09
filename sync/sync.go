package main

import "sync"

func main()  {
	mux := sync.Mutex{}
	mux.Unlock()
	//fatal error: sync: unlock of unlocked mutex
}
