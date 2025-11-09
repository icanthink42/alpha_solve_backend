env "local" {
  src = "file://schema.sql"
  url = "postgres://postgres@localhost:5432/alpha_solve?sslmode=disable"
  dev = "docker://postgres/15/dev?search_path=public"
}

env "prod" {
  src = "file://schema.sql"
  url = getenv("DATABASE_URL")
  dev = "docker://postgres/15/dev?search_path=public"
}

