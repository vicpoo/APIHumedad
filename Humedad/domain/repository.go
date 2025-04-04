// repository.go
package domain

import "github.com/vicpoo/APIHumedad/Humedad/domain/entities"

type TemperatureRepository interface {
	Save(temp entities.Temperature) error
	GetAll() ([]entities.Temperature, error)
}
