.PHONY: build install clean

build:
	go build -o saferay .

install: build
	./saferay install

clean:
	rm -f saferay
