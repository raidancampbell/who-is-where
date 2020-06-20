//usr/bin/env go run "$0" "$@" ; exit "$?"
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"net/http"
	"time"
)

const datasourceLocation = "./who-is-where.sqlite"

var db = Database{}

type Database struct {
	DB *gorm.DB
}

func main() {
	var err error
	db.DB, err = gorm.Open("sqlite3", datasourceLocation)
	if err != nil {
		panic(err)
	}

	db.DB.AutoMigrate(&LocationTable{})

	r := gin.Default()
	r.GET("/api/v1/:zone/:host", GormWrapper(db.DB, updateLocation))
	r.GET("/api/v1/:zone", GormWrapper(db.DB, dumpLocations))
	r.Run(":8081")
}

func dumpLocations(db *gorm.DB, c *gin.Context) {
	zone := c.Param("zone")
	fmt.Printf("received request to dump location from zone '%s'", zone)
	var locations []LocationTable

	subQuery := db.
		Model(&LocationTable{}).
		Select("location_tables.*").
		Where("zone = ?", zone).
		SubQuery()

	db.Select("location_tables.*").
		Joins("LEFT OUTER JOIN ? AS t1 on t1.host = location_tables.host AND location_tables.created_at < t1.created_at", subQuery, zone).
		Where("t1.host IS NULL").Find(&locations)

	b, err := json.Marshal(locations)
	if err != nil {
		http.Error(c.Writer, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	if _, err := c.Writer.Write(b); err != nil {
		http.Error(c.Writer, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

}

func updateLocation(db *gorm.DB, c *gin.Context) {
	host := c.Param("host")
	zone := c.Param("zone")
	fmt.Printf("received request for host '%s' in zone '%s' at time '%s'\n", host, zone, time.Now().String())
	db.Save(&LocationTable{
		Location: c.Request.RemoteAddr,
		Host:     host,
		Zone:     zone,
	})
}

// provides a closure to enable DB interactions
// given a DB and a function that needs (the DB plus whatever its original signature should be)
// this function creates a closure to return a function of the original signature, but with the DB in a closure.
func GormWrapper(db *gorm.DB, f func(db *gorm.DB, c *gin.Context)) func(c *gin.Context) {
	return func(c *gin.Context) {
		f(db, c)
	}
}

type LocationTable struct {
	gorm.Model `gorm:"-"`
	Location   string `gorm:"type:text"`
	Host       string `gorm:"type:text"`
	Zone       string `gorm:"type:text"`
}
