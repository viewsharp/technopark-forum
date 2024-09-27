migrate:
	goose postgres "postgres://postgres:postgres_password@localhost:5432/postgres" -dir migrations up
func:
	tech-db-forum func -u http://localhost:8000/api -r report.html
fill:
	tech-db-forum fill -u http://localhost:8000/api
perf:
	tech-db-forum perf -u http://localhost:8000/api
