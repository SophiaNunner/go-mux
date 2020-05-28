Travis-CI Build Status: [![Build Status](https://travis-ci.com/SophiaNunner/go-mux.svg?branch=master)](https://travis-ci.com/SophiaNunner/go-mux)

go-mux: Microservice in GoTutorial
Tutorial from semaphoreci.com


Prerequisites:
1. GitHub Account
2. PostgreSql installed and configured

- make sure user/role 'postgres' is available
- environment variables properly set
  - export APP_DB_USERNAME=postgres
  - export APP_DB_PASSWORD=whatever password you use
  - export APP_DB_NAME=postgres
- GoLang installed


Hints/Info:
- When you run go test it compiles all the files ending in _test.go into a test binary and then runs that binary to execute the tests. Since the go test binary is simply a compiled go program, it can process command line arguments like any other program
- 2nd part: add Travis (.travis.yml), see docs.travis-ci.com
