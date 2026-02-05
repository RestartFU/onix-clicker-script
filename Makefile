build:
	go build -trimpath -ldflags "-s -w -H=windowsgui" -o onix-clicker.exe cmd/main.go
