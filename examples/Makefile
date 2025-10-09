.PHONY: all clean read write create custom search dexport schema documentation bytes reader

all: clean read write create custom search dexport schema documentation bytes reader
read:
	cd read && go run ./read.go
write:
	cd write && go run write.go
create:
	cd create && go run create.go
custom:
	cd custom && go run custom.go
search:
	cd search && go run search.go
dexport:
	cd database && go run export.go
schema:
	cd schema && go run schema.go
documentation:
	cd documentation && go run documentation.go
bytes:
	cd bytes && go run bytes.go
reader:
	cd reader && go run reader.go
clean:
	cd read && rm -f debug.log
	cd write && rm -f debug.log
	cd create && rm -f debug.log rm -f TEST.DBF && rm -f TEST.FPT
	cd custom && rm -f debug.log
	cd search && rm -f debug.log
	cd database && rm -f debug.log
	cd schema && rm -f debug.log
	cd documentation && rm -f debug.log
	cd bytes && rm -f debug.log
	cd reader && rm -f debug.log
