.PHONY: build install clean run test lint fmt coverage

BINARY = dupclean
APP_ID = com.dupclean.app
INSTALL_PATH = /usr/local/bin/$(BINARY)

build:
	go build -o $(BINARY) .
	@echo "✅ Built: ./$(BINARY)"

install: build
	@echo "Installing dupclean CLI..."
	sudo mv $(BINARY) $(INSTALL_PATH)
	@echo "✅ Installed to $(INSTALL_PATH)"

uninstall:
	sudo rm -f $(INSTALL_PATH)
	@echo "🗑  Uninstalled CLI."

run:
	go run . $(FOLDER)

test:
	go test ./...

test-verbose:
	go test -v ./...

coverage:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

lint:
	golangci-lint run

fmt:
	go fmt ./...
	goimports -w -local dupclean .

vet:
	go vet ./...

clean:
	rm -f $(BINARY)
	rm -rf dist/
	rm -rf fyne-cross/
	rm -f coverage.out coverage.html

# Cross-compilation using fyne-cross (requires Docker)
cross: cross-linux cross-darwin cross-windows

cross-linux:
	@echo "Building Linux binaries with fyne-cross..."
	fyne-cross linux -app-id $(APP_ID)
	@echo "✅ Linux build: fyne-cross/dist/linux-arm64/"

cross-darwin:
	@echo "Building macOS binaries with fyne-cross..."
	fyne-cross darwin -app-id $(APP_ID)
	@echo "✅ macOS build: fyne-cross/dist/darwin-arm64/"

cross-windows:
	@echo "Building Windows binaries with fyne-cross..."
	fyne-cross windows -app-id $(APP_ID)
	@echo "✅ Windows build: fyne-cross/dist/windows-arm64/"

# Package builds into distribution-ready archives (runs cross first if needed)
package: cross package-only

package-only:
	@echo "Creating distribution packages..."
	@mkdir -p dist/packages dist/linux dist/darwin dist/windows

	@echo "Linux ARM64..."
	@xz -d fyne-cross/dist/linux-arm64/dupclean.tar.xz -c | tar -xf - -C fyne-cross/dist/linux-arm64
	@cp -r fyne-cross/dist/linux-arm64/usr/local/bin/dupclean dist/linux/dupclean-linux-arm64
	@tar -czf dist/packages/dupclean-linux-arm64.tar.gz -C dist/linux dupclean-linux-arm64

	@echo "macOS ARM64..."
	@cp -r fyne-cross/dist/darwin-arm64/dupclean.app dist/darwin/
	@tar -czf dist/packages/dupclean-darwin-arm64.tar.gz -C dist darwin/dupclean.app

	@echo "Windows ARM64..."
	@unzip -o fyne-cross/dist/windows-arm64/dupclean.exe.zip -d dist/windows/
	@mv dist/windows/dupclean.exe dist/windows/dupclean-windows-arm64.exe
	@zip -j -r dist/packages/dupclean-windows-arm64.zip dist/windows/dupclean-windows-arm64.exe

	@echo "✅ All packages created in dist/packages/"
	@ls -lh dist/packages/

# Build and package everything
release: cross-darwin-local package-only
	@echo "Creating DMG..."
	@hdiutil create -volname "DupClean" -srcfolder dist/darwin/dupclean.app -ov -format UDZO dist/packages/dupclean.dmg 2>/dev/null || true
	@echo "🚀 Release ready in dist/packages/"
	@ls -lh dist/packages/

# Build macOS locally (requires macOS)
cross-darwin-local:
	@echo "Building macOS locally (arm64)..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o dist/darwin/dupclean-darwin-arm64 .
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o dist/darwin/dupclean-darwin-amd64 .
	@echo "✅ Local macOS builds: dist/darwin/"
