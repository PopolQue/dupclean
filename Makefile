.PHONY: build install clean run

BINARY = dupclean
INSTALL_PATH = /usr/local/bin/$(BINARY)

build:
	go build -o $(BINARY) .
	@echo "✅ Built: ./$(BINARY)"

install: build
	sudo mv $(BINARY) $(INSTALL_PATH)
	@echo "✅ Installed to $(INSTALL_PATH)"
	@echo "   Run: dupclean <folder>"

uninstall:
	sudo rm -f $(INSTALL_PATH)
	@echo "🗑  Uninstalled."

run:
	go run . $(FOLDER)

clean:
	rm -f $(BINARY)
