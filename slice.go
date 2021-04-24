package main

import (
	"errors"
	"log"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
)

func sliceFile(rpath string, thisConfig config, outputFilepath string) error {
	//Get executable base on platform
	executable, err := selectBianry()
	if err != nil {
		return err
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

	//Inject the output filepath to the slicing paramter
	slicingParamters = append(slicingParamters, "--output="+outputFilepath)

	cmd := exec.Command(executable, slicingParamters...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(string(out))
		return err
	}

	//OK
	return nil
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
