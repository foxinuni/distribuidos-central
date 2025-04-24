package main

type Config struct {
	Faculties int
	Address   string
}

var Faculties = []string{
	"Ingeniería",
	"Arquitectura",
	"Derecho",
	"Economía",
	"Educación",
	"Medicina",
	"Psicología",
	"Ciencias Sociales",
	"Derecho",
	"Administracion",
}

var Programs = map[int][]string{
	0: {"Ingeniería Civil", "Ingeniería de Sistemas", "Ingeniería Industrial", "Ingeniería Electrónica", "Ingeniería Eléctrica"},
	1: {"Arquitectura", "Diseño de Interiores", "Urbanismo", "Arquitectura Técnica", "Restauración"},
	2: {"Derecho Civil", "Derecho Penal", "Derecho Mercantil", "Derecho Internacional", "Derecho Constitucional"},
	3: {"Economía General", "Economía Financiera", "Economía Internacional", "Economía Laboral", "Economía Agraria"},
	4: {"Educación Primaria", "Educación Secundaria", "Educación Especial", "Educación Física", "Educación Infantil"},
	5: {"Medicina General", "Medicina Interna", "Medicina Pediátrica", "Medicina Geriátrica", "Medicina Familiar"},
	6: {"Psicología Clínica", "Psicología Educativa", "Psicología Laboral", "Psicología Social", "Psicología Experimental"},
	7: {"Sociología General", "Antropología", "Ciencias Políticas", "Trabajo Social", "Relaciones Internacionales"},
	8: {"Derecho Corporativo", "Derecho Laboral", "Derecho Ambiental", "Derecho de Familia", "Derecho Administrativo"},
	9: {"Administración de Empresas", "Marketing", "Finanzas", "Recursos Humanos", "Logística"},
}
