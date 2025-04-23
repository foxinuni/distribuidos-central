package main

type ServerOptions func(*Server)

func WithPort(port int) ServerOptions {
	return func(c *Server) {
		c.port = port
	}
}

func WithWorkerCount(workerCount int) ServerOptions {
	return func(c *Server) {
		c.workers = workerCount
	}
}
