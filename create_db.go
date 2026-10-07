package main
import (
    "database/sql"
    "log"
    _ "github.com/jackc/pgx/v5/stdlib"
)
func main() {
    db, err := sql.Open("pgx", "postgres://postgres:@127.0.0.1:5432/postgres?sslmode=disable")
    if err != nil { log.Fatal(err) }
    defer db.Close()
    _, err = db.Exec("CREATE DATABASE einvoice_saas")
    if err != nil {
        log.Println("Info:", err)
    } else {
        log.Println("Base de données 'einvoice_saas' créée avec succès.")
    }
}