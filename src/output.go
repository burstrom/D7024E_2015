package dht

import (
	//"fmt"
	"github.com/fatih/color"
)

//
func Error(s string) {
	Red := color.New(color.Bold, color.FgRed).PrintFunc()
	Red(s)
}
func Errorln(s string) {
	Error(s + "\n")
}

func Notice(s string) {
	Green := color.New(color.Bold, color.FgGreen).PrintFunc()
	Green(s)
}
func Noticeln(s string) {
	Notice(s + "\n")
}

func Info(s string) {
	Info := color.New(color.Bold, color.FgMagenta).PrintFunc()
	Info(s)
}
func Infoln(s string) {
	Info(s + "\n")
}

func Warn(s string) {
	Info := color.New(color.Bold, color.FgYellow).PrintFunc()
	Info(s)
}
func Warnln(s string) {
	Warn(s + "\n")
}

/*
Red := color.New(color.FgRed).PrintFunc()
Red("Warning")
Red("Error: %s", err)

Notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
Notice("Don't frget this...")
*/
