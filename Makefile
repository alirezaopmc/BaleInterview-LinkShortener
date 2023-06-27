run:
	docker compose up -d
	sleep 2 #better to check docker status
	go run .

stop:
	docker compose down
