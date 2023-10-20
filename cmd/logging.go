package cmd

import (
	"log"
)

func LogInfo(format string, v ...any) {
	log.Printf(Blue+format+Reset, v...)
}

func LogError(format string, v ...any) {
	log.Printf(Red+format+Reset, v...)
}
