package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
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

	//Get executable base on platform
	executable, err := selectBianry()
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Build the slicing paramter
	slicingParamters := []string{rpath} //The first paramter is the stl filepath itself

	slicingParamters = append(slicingParamters, "--hot-end-temperature="+strconv.Itoa(thisConfig.HotEndTemperature))
	slicingParamters = append(slicingParamters, "--bed-temperature="+strconv.Itoa(thisConfig.BedTemperature))

	//Calculate the center of the bed
	bedCenterX := int(thisConfig.BedWidth * 1000 / 2)
	bedCenterY := int(thisConfig.BedDepth * 1000 / 2)
	slicingParamters = append(slicingParamters, "--center="+strconv.Itoa(bedCenterX)+"_"+strconv.Itoa(int(bedCenterY))+"_0")

	slicingParamters = append(slicingParamters, "--extrusion-width="+strconv.Itoa(thisConfig.ExtrusionWidth))
	slicingParamters = append(slicingParamters, "--filament-diameter="+strconv.Itoa(thisConfig.FilamentDiameter))
	slicingParamters = append(slicingParamters, "--layer-thickness="+strconv.Itoa(thisConfig.LayerThickness))
	slicingParamters = append(slicingParamters, "--extrusion-multiplier="+strconv.Itoa(thisConfig.ExtrusionMultiplier))
	slicingParamters = append(slicingParamters, "--layer-speed="+strconv.Itoa(thisConfig.LayerSpeed))

	slicingParamters = append(slicingParamters, "--move-speed="+strconv.Itoa(thisConfig.MoveSpeed))
	slicingParamters = append(slicingParamters, "--number-top-layers="+strconv.Itoa(thisConfig.NumberTopLayers))
	slicingParamters = append(slicingParamters, "--number-bottom-layers="+strconv.Itoa(thisConfig.NumberBottomLayers))
	slicingParamters = append(slicingParamters, "--brim-count="+strconv.Itoa(thisConfig.BrimCount))
	slicingParamters = append(slicingParamters, "--skirt-count="+strconv.Itoa(thisConfig.SkirtCount))

	//Initial layer settings
	slicingParamters = append(slicingParamters, "--initial-bed-temperature="+strconv.Itoa(thisConfig.InitialBedTemperature))
	slicingParamters = append(slicingParamters, "--initial-hot-end-temperature="+strconv.Itoa(thisConfig.InitialHotEndTemperature))
	slicingParamters = append(slicingParamters, "--initial-layer-speed="+strconv.Itoa(thisConfig.InitialLayerSpeed))
	slicingParamters = append(slicingParamters, "--initial-layer-thickness="+strconv.Itoa(thisConfig.InitialLayerThickness))

	//Use support
	if thisConfig.SupportEnabled {
		slicingParamters = append(slicingParamters, "--support-enabled")
	}

	//Execute the slicing to tmp folder
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

	//Inject the output filepath to the slicing paramter
	slicingParamters = append(slicingParamters, "--output="+outputFilepath)

	cmd := exec.Command(executable, slicingParamters...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		sendErrorResponse(w, err.Error())
	}
	//OK
	log.Println(string(out))

	//Return the filename for preview
	js, _ := json.Marshal(filepath.ToSlash(filepath.Join("tmp:/SlicerA/", outputFilename)))
	sendJSONResponse(w, string(js))

}

func selectBianry() (string, error) {
	binaryName := "./goslice/goslice"
	binaryExecPath := ""
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "arm" {
			binaryExecPath = binaryName + "-linux-arm.elf"
		} else if runtime.GOARCH == "arm64" {
			binaryExecPath = binaryName + "-linux-arm64.elf"
		} else if runtime.GOARCH == "386" {
			binaryExecPath = binaryName + "-linux-386.elf"
		} else if runtime.GOARCH == "amd64" {
			binaryExecPath = binaryName + "-linux-amd64.elf"
		}

		if binaryExecPath != "" {
			//Set it to absolute for cross platform compeibility
			abspath, _ := filepath.Abs(binaryExecPath)
			return abspath, nil
		} else {
			return "", errors.New("Platform not supported")
		}

	} else if runtime.GOOS == "windows" {
		binaryExecPath = binaryName + "-windows-amd64.exe"
		abspath, _ := filepath.Abs(binaryExecPath)
		return abspath, nil
	} else {
		//Not supported
		return "", errors.New("Platform not supported")
	}
}
