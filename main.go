// entry point for our application
package main

import "os"

func main() {
	a := App{}
	a.Initialize( // This assumes that you use environment variables APP_DB_USERNAME, APP_DB_PASSWORD, and APP_DB_NAME to store your database’s username, password, and name respectively.
		os.Getenv("APP_DB_USERNAME"),
		os.Getenv("APP_DB_PASSWORD"),
		os.Getenv("APP_DB_NAME"))

	a.Run(":8010")
}
