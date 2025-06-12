package database

import "os"

var (
	Host     = os.Getenv("POSTGRES_HOST")
	Username = os.Getenv("POSTGRES_USER")
	Password = os.Getenv("POSTGRES_PASSWORD")
	Dbname   = os.Getenv("POSTGRES_DB")
)
