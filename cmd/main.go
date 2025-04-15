package main

import (
	"fmt"
	"log"

	debianbuild "bangun/internal/build/usecase"
)

func init() {
	log.SetFlags(log.Flags() | log.Llongfile)
}

func main() {
	fmt.Println("bangun")

	debianBuild := debianbuild.NewDebianBuildUsecase()

	debianBuild.Build(debianbuild.DebianBuildRequest{
		Distro: "noble",
		Arch:   "amd64",
		GitURL: "https://github.com/herpiko/cubetimer.git",
	})
}
