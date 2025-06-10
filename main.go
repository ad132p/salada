package main

import (
	"fmt"
	"net/http"
	"os"
	"salada/internal/blog/controller"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.Static("/assets/", "./web/assets")
	router.Static("/images/", "./web/images")
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

	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"title": "404",
		})
	})

	router.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"code": "METHOD_NOT_ALLOWED", "message": "405 method not allowed"})
	})

	controller.RegisterBlogRoutes(router)

	bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))
	router.Run(bindIp)
}
