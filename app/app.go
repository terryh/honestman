package app

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	// _ "github.com/go-sql-driver/mysql"
)

var (
	Version     = "0.0.1"
	PackageName = "Honestman"
	LastUpdated = "2018/03/01"
	Authors     = `Terry Huang`

	// share in application
	App    *Context
	debug  bool
	dbName = ""
	dbHost = ""
	dbUser = ""
	// not use in this DEMO
	dbPass = ""
	port   = ""
)

// Context
type Context struct {
	DB    *sqlx.DB
	Port  string
	Debug bool
}

// ContextInit for initialize
func ContextInit(dburi string, port string, debug bool) *Context {
	// var real name space
	App = new(Context)

	// db
	db, err := sqlx.Open("postgres", dburi)

	if err != nil {
		log.Fatalln(err)
	}

	App.DB = db
	App.Port = port
	App.Debug = debug
	return App
}

func init() {
	// Give the default value here

	flag.StringVar(&dbName, "dbname", "honest", `database name to conect`)
	flag.StringVar(&dbHost, "dbhost", "localhost", `database host to conec`)
	flag.StringVar(&dbUser, "dbuser", "terry", `database user for connection`)
	flag.StringVar(&dbPass, "dbpass", "", `database password for connection`)
	flag.StringVar(&port, "port", ":3000", `address for listen default is :3000`)
	flag.BoolVar(&debug, "debug", false, `Flag for DEBUG, Default is: false`)

	flag.Parse()
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if os.Getenv("DBNAME") != "" {
		dbName = os.Getenv("DBNAME")
	}

	if os.Getenv("DBHOST") != "" {
		dbHost = os.Getenv("DBHOST")
	}

	if os.Getenv("DBUSER") != "" {
		dbUser = os.Getenv("DBUSER")
	}

	if os.Getenv("DBPASS") != "" {
		dbPass = os.Getenv("DBPASS")
	}

	if os.Getenv("PORT") != "" {
		port = os.Getenv("PORT")
	}

	if os.Getenv("DEBUG") != "" {
		debug = true
	}

	if dbName == "" {
		printhelp()
		os.Exit(0)
	}

	fmt.Println(PackageName, Version)
	fmt.Printf("database: %s@%s %s\n", dbUser, dbHost, dbName)
	if debug {
		fmt.Println("Running in DEBUG mode")
	}
}

func printhelp() {
	fmt.Println("Name:", PackageName, Version)
	fmt.Println("Usage:")
	flag.PrintDefaults()
}

func NewContext() *Context {
	dbURI := fmt.Sprintf(" dbname=%s host=%s user=%s sslmode=disable", dbName, dbHost, dbUser)
	return ContextInit(dbURI, port, debug)
}
