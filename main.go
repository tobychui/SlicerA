package main

/*
	SlicerA
	The basic 3D model slicer for 3D printer based on GoSlice open souce project

	Author: tobychui
*/

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"imuslab.com/SlicerA/mod/aroz"
)

var (
	handler *aroz.ArozHandler
)

//Kill signal handler. Do something before the system the core terminate.
func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Shutting down SlicerA module")

		//Clean up the tmp folder if it exists
		if fileExists("./tmp") {
			os.RemoveAll("./tmp")
		}
		os.Exit(0)
	}()
}

func main() {
	//Start the aoModule pipeline (which will parse the flags as well). Pass in the module launch information
	handler = aroz.HandleFlagParse(aroz.ServiceInfo{
		Name:         "SlicerA",
		Desc:         "A basic STL 3D model slicer for the ArozOS Cloud Platform",
		Group:        "Utilities",
		IconPath:     "SlicerA/img/small_icon.png",
		Version:      "0.3.6", //Try to match the GoSlice version we are using
		StartDir:     "SlicerA/index.html",
		SupportFW:    true,
		LaunchFWDir:  "SlicerA/index.html",
		SupportEmb:   true,
		LaunchEmb:    "SlicerA/index.html",
		InitFWSize:   []int{1060, 670},
		InitEmbSize:  []int{1060, 670},
		SupportedExt: []string{".stl"},
	})

	//Register the standard web services urls
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)

	//Handle the slicing process
	http.HandleFunc("/slice", handleSlicing)
	http.HandleFunc("/sliced", handleSliceAndDispose)
	http.HandleFunc("/saveGcode", handleSaveGcode)

	//Setup the close handler to handle Ctrl+C on terminal
	SetupCloseHandler()

	//Any log println will be shown in the core system via STDOUT redirection. But not STDIN.
	log.Println("SlicerA started. Listening on " + handler.Port)
	err := http.ListenAndServe(handler.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
