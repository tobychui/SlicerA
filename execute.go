package main

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type config struct {
	HotEndTemperature        int     `json:"hot-end-temperature"`
	BedTemperature           int     `json:"bed-temperature"`
	BedWidth                 float64 `json:"bed-width"`
	BedDepth                 float64 `json:"bed-depth"`
	ExtrusionWidth           int     `json:"extrusion-width"`
	FilamentDiameter         int     `json:"filament-diameter"`
	LayerThickness           int     `json:"layer-thickness"`
	ExtrusionMultiplier      int     `json:"extrusion-multiplier"`
	LayerSpeed               int     `json:"layer-speed"`
	MoveSpeed                int     `json:"move-speed"`
	NumberTopLayers          int     `json:"number-top-layers"`
	NumberBottomLayers       int     `json:"number-bottom-layers"`
	BrimCount                int     `json:"brim-count"`
	SkirtCount               int     `json:"skirt-count"`
	SupportEnabled           bool    `json:"support-enabled"`
	InitialBedTemperature    int     `json:"initial-bed-temperature"`
	InitialHotEndTemperature int     `json:"initial-hot-end-temperature"`
	InitialLayerSpeed        int     `json:"initial-layer-speed"`
	InitialLayerThickness    int     `json:"initial-layer-thickness"`
}

//Save the tmp gcode to actual file location
func handleSaveGcode(w http.ResponseWriter, r *http.Request) {
	src, err := mv(r, "src", true)
	if err != nil {
		sendErrorResponse(w, "Invalid src path")
		return
	}

	dest, err := mv(r, "dest", true)
	if err != nil {
		sendErrorResponse(w, "Invalid dest path")
		return
	}

	//Convert the src and dest vpath to rpath
	realSourcePath, err := resolveVirtualPath(w, r, src)
	if err != nil {
		sendErrorResponse(w, "Invalid src path")
		return
	}

	realDestPath, err := resolveVirtualPath(w, r, dest)
	if err != nil {
		sendErrorResponse(w, "Invalid dest path")
		return
	}

	//Copy the file from source to dest
	if fileExists(realSourcePath) {
		//Copy the file
		gcode, err := ioutil.ReadFile(realSourcePath)
		if err != nil {
			sendErrorResponse(w, "Unable to read tmp gcode file")
			return
		}

		err = ioutil.WriteFile(realDestPath, gcode, 0755)
		if err != nil {
			sendErrorResponse(w, "Unable to write to destination file")
			return
		}

		sendOK(w)
	} else {
		sendErrorResponse(w, "Source gcode file not exists")
		return
	}
}

//Slice and return the results, dispose source
func handleSliceAndDispose(w http.ResponseWriter, r *http.Request) {
	options, err := mv(r, "options", false)
	if err != nil {
		sendErrorResponse(w, "Invalid options given")
		return
	}

	stlContent, err := ioutil.ReadAll(r.Body)
	if err != nil {
		sendErrorResponse(w, "Unable to read stl content")
		return
	}

	//Create a tmp folder
	os.MkdirAll(*tmpFolderPath, 0755)

	//Buffer the received base64 stl file into the tmp folder
	noheaderContent := strings.SplitN(string(stlContent), ",", 2)[1]
	dec, err := base64.StdEncoding.DecodeString(noheaderContent)
	if err != nil {
		sendErrorResponse(w, "Unable to read stl content")
		return
	}

	//Write to file
	fileID := strconv.Itoa(int(time.Now().Unix()))
	tmpFilepath := filepath.Join(*tmpFolderPath, fileID+".stl")
	outFilepath := filepath.Join(*tmpFolderPath, fileID+".gcode")

	err = ioutil.WriteFile(tmpFilepath, dec, 0755)
	if err != nil {
		sendErrorResponse(w, "Write file to buffer folder failed")
		return
	}

	//Parse the config
	thisConfig := config{}
	err = json.Unmarshal([]byte(options), &thisConfig)
	if err != nil {
		log.Println(err.Error())
		sendErrorResponse(w, "Parse option failed")
		return
	}

	//Slice the file
	err = sliceFile(tmpFilepath, thisConfig, outFilepath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Get the output
	finalGcode, err := ioutil.ReadFile(outFilepath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Remove the tmp files
	os.Remove(tmpFilepath)
	os.Remove(outFilepath)

	//Send back to sliced Gcode
	sendTextResponse(w, string(finalGcode))
}

//Handle slicing request and return the gcode
func handleSlicing(w http.ResponseWriter, r *http.Request) {
	//Get slicing profile
	options, err := mv(r, "options", true)
	if err != nil {
		sendErrorResponse(w, "Invalid options given")
		return
	}

	//Get the src file
	vpath, err := mv(r, "file", true)
	if err != nil {
		sendErrorResponse(w, "Invalid input file given")
		return
	}

	//Convert the input vpath to realpath
	rpath, err := resolveVirtualPath(w, r, vpath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Parse the option into the required structure
	thisConfig := config{}
	err = json.Unmarshal([]byte(options), &thisConfig)
	if err != nil {
		log.Println(err.Error())
		sendErrorResponse(w, "Parse option failed")
		return
	}

	//Get the output filename from ArozOS
	tmpFolderAbsolutePath, err := resolveVirtualPath(w, r, "tmp:/SlicerA/")
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Create the target tmp folder if not exists
	os.MkdirAll(tmpFolderAbsolutePath, 0755)

	//Generate the name of the tmp gcode file
	outputFilename := strconv.Itoa(int(time.Now().Unix())) + ".gcode"
	outputFilepath := filepath.Join(tmpFolderAbsolutePath, outputFilename)

	err = sliceFile(rpath, thisConfig, outputFilepath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Return the filename for preview
	js, _ := json.Marshal(filepath.ToSlash(filepath.Join("tmp:/SlicerA/", outputFilename)))
	sendJSONResponse(w, string(js))

}
