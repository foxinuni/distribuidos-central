package handler

type RouteHandler func(interface{}) (interface{}, error)

func (s *Server) registerRoutes() {
	s.routes["health-check"] = s.healthCheckController.HealthCheck
	s.routes["allocate"] = s.allocationsController.Allocate
	s.routes["confirm"] = s.allocationsController.Confirm
}
