package server

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
)

func Run(router *gin.Engine) {
	mode := os.Getenv("MODE")
	if mode != "dev" && mode != "prod" {
		log.Fatalf("invalid MODE %q: must be \"dev\" or \"prod\"", mode)
	}

	// 'dev' mode run on :8080, and prod deploys salada.dev
	if mode == "dev" {
		bindIp := fmt.Sprintf("%s:8080", os.Getenv("BIND_IP"))
		gin.SetMode(gin.DebugMode) 
		router.RunTLS(bindIp, "cert.pem", "key.pem")
	} else {
		m := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("www.salada.dev", "salada.dev"),
			Cache:      autocert.DirCache("/var/www/.cache"),
		}

		log.Fatal(autotls.RunWithManager(router, &m))
	}
}
