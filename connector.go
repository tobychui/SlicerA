package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

/*
	Helper.go

	This section of code handle the connection between the ArozOS platform API and the SlicerA file IO
	For the API scripting, see AJGI Documentation.md in the src folder of the ArozOS project github page
*/

//resolveVirtualPath get the absolute path from the given vpath. Return absolutepath and error if any
func resolveVirtualPath(w http.ResponseWriter, r *http.Request, vpath string) (string, error) {
	//Get username and token from request
	_, token := handler.GetUserInfoFromRequest(w, r)

	//Create an AGI Call that get the user desktop files
	script := `
		sendResp((decodeAbsoluteVirtualPath("` + vpath + `")).split("\\").join("/"));	
	`

	//Execute the AGI request on server side
	resp, err := handler.RequestGatewayInterface(token, script)
	if err != nil {
		//Something went wrong when performing POST request
		log.Println(err)
	} else {
		//Try to read the resp body
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		return filepath.ToSlash(string(bodyBytes)), nil
	}

	return "", errors.New("Unknown error occured")
}
