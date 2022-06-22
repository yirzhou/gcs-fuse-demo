default: run

build:
	go build main.go

run:
	go run main.go

umount:
	umount test_gcs
