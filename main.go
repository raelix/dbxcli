// Copyright © 2016 Dropbox, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dropbox/dbxcli/cmd"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("dbxcli version:", version)
		sdkVersion, specVersion := dropbox.Version()
		fmt.Println("SDK version:", sdkVersion)
		fmt.Println("Spec version:", specVersion)
	},
}

func init() {
	// Log date, time and file information by default
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	cmd.RootCmd.AddCommand(versionCmd)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func main() {
	// cmd.Execute()
	r := gin.Default()
	r.GET("/", root)
	r.GET("/backup", backup)
	r.Run(fmt.Sprintf(":%s", getEnv("HTTP_PORT", "8000")))
}

func backup(c *gin.Context) {
	err := cmd.Init()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
	} else {
		src := c.DefaultQuery("src", "/tmp")
		isFile, err := isFile(src)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "the file does not exist",
			})
			return
		}
		if !isFile {
			tarAndUpload(c, src)
		} else {
			uploadFile(c, src)
		}

	}
}

func tarAndUpload(c *gin.Context, src string) {
	dir := filepath.Dir(src)
	parent := filepath.Base(dir)
	dst := fmt.Sprintf("/tmp/%s.tar", parent)
	remoteDst := fmt.Sprintf("/%s", filepath.Base(dst))
	err := cmd.Tar(src, dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "tar failed",
		})
		return
	}
	go cmd.GenericPut(false, dst, remoteDst, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "the tarball upload will continue in background",
	})
}

func uploadFile(c *gin.Context, src string) {
	dst := filepath.Base(src)
	cmd.GenericPut(false, src, fmt.Sprintf("/%s", dst), false)
	c.JSON(http.StatusOK, gin.H{
		"message": "upload file completed",
	})
}

func isFile(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return false, nil
	} else {
		return true, nil
	}
}

func root(c *gin.Context) {
	err := cmd.Init()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
	} else {
		res, err := cmd.GenericAccount()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "not connected",
			})
		} else {
			c.JSON(http.StatusOK, res)
		}
	}
}
