package handlers

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/render"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func StatusHandler(g *gin.Context) {
	// TODO: Посмотрите вывод в консоль этих двух строк
	// fmt.Printf("Response: \n%+v\n\n", w)
	// fmt.Printf("Request: \n%+v\n\n", r)

	slog.Info("Received health check request", slog.String("method", gin.Mode()), slog.String("url", g.Request.RequestURI))
	render.JSON(g.Writer, g.Request, HealthResponse{Status: "OK"})
}
