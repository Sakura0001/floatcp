package main

import (
	"github.com/mweb/floatcompare"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(floatcompare.NewAnalyzer())
}
