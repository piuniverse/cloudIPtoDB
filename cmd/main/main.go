package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github/stclaird/cloudIPtoDB/pkg/config"
	"github/stclaird/cloudIPtoDB/pkg/ipfile"
	"github/stclaird/cloudIPtoDB/pkg/ipnet"
	"github/stclaird/cloudIPtoDB/pkg/models"

	_ "github.com/mattn/go-sqlite3"
)

var confObj config.Config
var db *sql.DB

func init() {
	confObj = config.NewConfig()

	if confObj.Downloaddir == "" {
		confObj.Downloaddir = "downloadedfiles"
	}

	err := os.MkdirAll(confObj.Downloaddir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	os.MkdirAll(confObj.Dbdir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	full_path := fmt.Sprintf("%s/%s", confObj.Dbdir, confObj.Dbfile)
	file, err := os.Create(full_path)

	if err != nil {
		log.Println("Os Create Error: ", err)
	}

	file.Close()

	models.DB, _ = sql.Open("sqlite3", full_path)
	if err != nil {
		log.Fatal(err)
	}

	models.SetupDB(models.DB)
	db = models.DB
}

func main() {

	for _, i := range confObj.Ipfiles {

		var cidrs []string

		downloadto := fmt.Sprintf("%s/%s", confObj.Downloaddir, i.DownloadFileName)

		fmt.Println(downloadto)
		fmt.Println("downloaddir", confObj.Downloaddir)

		var url string
		url = i.Url
		fmt.Println(url)

		switch i.Cloudplatform {
		case "azure":
			url = ipfile.ResolveAzureDownloadUrl() //azure download file changes so we need to figure out what the latest path is
		}

		var FileObj ipfile.IpfileTXT
		FileObj.Download(downloadto, url)
		cidrs_raw := ipfile.AsText[ipfile.IpfileTXT](downloadto)
		cidrs = FileObj.Process(cidrs_raw)

		for _, cidr := range cidrs {
			processedCidr, err := ipnet.PrepareCidrforDB(cidr)
			if err != nil {
				fmt.Println("Error: ", err)
			}

			if processedCidr.Iptype == "IPv4" {
				c := models.CidrObject{
					Net:           cidr,
					Start_ip:      processedCidr.NetIPDecimal,
					End_ip:        processedCidr.BcastIPDecimal,
					Url:           i.Url,
					Cloudplatform: i.Cloudplatform,
					Iptype:        processedCidr.Iptype,
				}

				models.AddCidr(db, c)

			}

		}
	}

}
