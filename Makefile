dev-prep:
	sudo mkdir /usr/share/bangun
	sudo ln -s $(shell pwd)/scripts /usr/share/bangun/scripts

run:
	go run cmd/main.go
	go run cmd/main.go
