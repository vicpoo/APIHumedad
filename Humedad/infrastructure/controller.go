// controller.go
package infrastructure

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vicpoo/APIHumedad/Humedad/application"
)

type TemperatureController struct {
	tempUseCase *application.TemperatureUseCase
}

func NewTemperatureController(tempUseCase *application.TemperatureUseCase) *TemperatureController {
	return &TemperatureController{
		tempUseCase: tempUseCase,
	}
}

func (tc *TemperatureController) GetAllTemperatures(c *gin.Context) {
	temps, err := tc.tempUseCase.GetAllTemperatures()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, temps)
}
