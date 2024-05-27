package main

import (
	"github.com/Sakura0001/floatcp"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(floatcp.NewAnalyzer())
}
