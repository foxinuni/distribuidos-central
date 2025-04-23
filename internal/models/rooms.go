package models

type SolicitudAulas struct {
	Facultad     string `json:"facultad"`
	Aulas        int    `json:"aulas"`
	Laboratorios int    `json:"laboratorios"`
}

type RespuestaAulas struct {
	Aulas []struct {
		Edificio int  `json:"edificio"`
		Numero   int  `json:"numero"`
		Adaptado bool `json:"adaptado"`
	} `json:"aulas"`

	Laboratorios []struct {
		Edificio int `json:"edificio"`
		Numero   int `json:"numero"`
	} `json:"laboratorios"`
}
