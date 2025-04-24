package controllers

type HealthCheckController struct{}

func NewHealthCheckController() *HealthCheckController {
	return &HealthCheckController{}
}

func (c *HealthCheckController) HealthCheck(_ interface{}) (interface{}, error) {
	return "OK", nil
}
