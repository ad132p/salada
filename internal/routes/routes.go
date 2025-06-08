package routes

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetRouter() {
	router := gin.Default()

	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
	router.GET("/pictures", getPictures)
	router.POST("/pictures", postPictures)

	router.LoadHTMLGlob("web/templates/html/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Home",
		})
	})
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", gin.H{
			"title": "About",
		})
	})

	router.GET("/contact", func(c *gin.Context) {
		c.HTML(http.StatusOK, "contact.html", gin.H{
			"title": "Contact",
		})
	})

	router.GET("/blog", func(c *gin.Context) {
		c.HTML(http.StatusOK, "blog.html", gin.H{
			"title": "Contact",
		})
	})

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"title": "404",
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	})

	bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))

	router.Run(bindIp)
}

type picture struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Strain    string `json:"artist"`
	Timestamp int    `json:"time"`
}

func postPictures(c *gin.Context) {
	var newPicture picture

	if err := c.BindJSON(&newPicture); err != nil {
		return
	}

	// Add the new picture to the slice.
	pictures = append(pictures, newPicture)
	c.IndentedJSON(http.StatusCreated, newPicture)
}

func getPictures(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, pictures)
}

var pictures = []picture{
	{ID: "1", Title: "vegging till when?", Strain: "Northern Lights", Timestamp: 1748392304},
}
