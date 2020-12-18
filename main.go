package main

import "log"

func main() {
	log.Println("Starting TCP proxy...")
	newProxy(":6061", ":6062").listen()
}
