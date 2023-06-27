run:
	docker compose up -d
	sleep 5 #better to check docker status
	go run .

stop:
	docker compose down
