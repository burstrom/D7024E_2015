package dht

import (
	//"fmt"
	"github.com/fatih/color"
)

//
func Error(s string) {
	Red := color.New(color.FgRed).PrintFunc()
	Red(s)
}

/*
Red := color.New(color.FgRed).PrintFunc()
Red("Warning")
Red("Error: %s", err)

Notice := color.New(color.Bold, color.FgGreen).PrintlnFunc()
Notice("Don't frget this...")
*/
