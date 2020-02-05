package cmd

import (
	"cf-tool/client"
	"os"
	"path/filepath"
)

// Status command
func Status(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cln := client.New(currentPath)
	username := args["<username>"].(string)

	return cln.SaveStatus(username, filepath.Join(currentPath, "data"))
}

// UserStatuses
func UserStatuses(args map[string]interface{}) error {
	currentPath, err := os.Getwd()
	if err != nil {
		return err
	}
	cln := client.New(currentPath)
	var handles string
	if args["[<handles>]"] != nil {
		handles = args["[<handles>]"].(string)
	}
	if handles == "" {
		handles = filepath.Join(currentPath, "data/handles.json")
	}
	return cln.SaveUserStatuses(handles, filepath.Join(currentPath, "data"))
}
