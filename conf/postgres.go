package conf

type postgresStruct struct {
	User     string
	Password string
	DBName   string
	Port     string
}

var Postgres postgresStruct

func init() {
	Postgres = postgresStruct{
		User:     "docker",
		Password: "docker",
		DBName:   "docker",
		Port:     "5432",
	}
}
