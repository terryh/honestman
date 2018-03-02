### DB migration
install goose go get -u github.com/pressly/goose/cmd/goose

### Show status
goose postgres "user=terry dbname=honest host=localhost sslmode=disable" status

### Create New migration
goose postgres "user=terry dbname=honest host=localhost sslmode=disable" create your_new_migration sql

### Upgarde db schema
goose postgres "user=terry dbname=honest host=localhost sslmode=disable" up

### Down to ? version
goose postgres "user=terry dbname=honest host=localhost sslmode=disable" down-to ?

### Please read help
goose --help

