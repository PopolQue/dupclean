package gui

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// updateContent updates the content container (preserves sidebar)
func (state *AppState) updateContent(content fyne.CanvasObject) {
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.ContentContainer != nil {
		state.ContentContainer.Objects = []fyne.CanvasObject{content}
		state.ContentContainer.Refresh()
	}
}

func RunGUI() {
	log.Println("RunGUI: starting...")

	fyneApp := app.NewWithID("com.popolque.dupclean")
	fyneApp.Settings().SetTheme(NewDupCleanTheme())

	log.Println("RunGUI: setting icon...")
	fyneApp.SetIcon(appLogo)

	log.Println("RunGUI: creating window...")
	w := fyneApp.NewWindow("DupClean - Duplicate File Finder & Cache Cleaner")
	w.Resize(fyne.NewSize(1200, 800))

	log.Println("RunGUI: creating states...")
	pm := NewProcessManager()
	dupState := &AppState{
		Window:            w,
		ProcessManager:    pm,
		FolderPath:        "",
		ScanAll:           false,
		IncludeHidden:     false,
		FollowSymlinks:    false,
		SimilarityPct:     90,
		Depth:             2,
		IsScanning:        false,
		ProgressText:      "Ready",
		ProgressValue:     0,
		Groups:            nil,
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
		playerDone:        make(chan struct{}, 1), // Buffered to prevent goroutine leak
	}

	cacheState := NewCacheCleanerState(w, pm, 0)
	diskState := NewDiskAnalyzerState(w, pm)

	log.Println("RunGUI: creating main layout with sidebar...")
	w.SetContent(createMainLayoutWithSidebar(dupState, cacheState, diskState))

	// Set up cleanup on window close
	w.SetOnClosed(func() {
		log.Println("RunGUI: cleaning up...")
		stopPlayback(dupState)
	})

	log.Println("RunGUI: showing window...")
	ShowChangelogIfNeeded(w)
	w.ShowAndRun()
	log.Println("RunGUI: window closed")
}

// createMainLayoutWithSidebar creates the main application layout with sidebar navigation
func createMainLayoutWithSidebar(dupState *AppState, cacheState *CacheCleanerState, diskState *DiskAnalyzerState) fyne.CanvasObject {
	// App header
	logo := canvas.NewImageFromResource(appLogo)
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(48, 48))

	appName := widget.NewLabel("DupClean")
	appName.TextStyle = fyne.TextStyle{Bold: true}
	appName.SizeName = theme.SizeNameHeadingText

	appSubtitle := widget.NewLabel("All-in-one disk cleanup tool")
	appSubtitle.TextStyle = fyne.TextStyle{Italic: true}
	appSubtitle.SizeName = theme.SizeNameSubHeadingText

	header := container.NewPadded(container.NewHBox(
		logo,
		container.NewVBox(appName, appSubtitle),
		layout.NewSpacer(),
	))

	// Create content area
	contentContainer := container.NewStack()

	// Create view widgets
	generalFinderView := DuplicateFinderWidget(dupState, "byte")
	audioFinderView := DuplicateFinderWidget(dupState, "audio")
	photoFinderView := DuplicateFinderWidget(dupState, "photo")
	cacheCleanerView := CacheCleanerWidget(cacheState)
	diskAnalyzerView := DiskAnalyzerWidget(diskState)

	updaterState := NewUpdaterState(dupState.Window)
	updaterView := UpdaterWidget(updaterState)

	// Store content container reference in states
	dupState.ContentContainer = contentContainer
	cacheState.ContentContainer = contentContainer
	diskState.ContentContainer = contentContainer

	// Navigation callback - updates content container
	onNavigate := func(viewIndex int) {
		log.Printf("Navigating to view index: %d", viewIndex)
		switch viewIndex {
		case 0:
			contentContainer.Objects = []fyne.CanvasObject{generalFinderView}
		case 1:
			contentContainer.Objects = []fyne.CanvasObject{audioFinderView}
		case 2:
			contentContainer.Objects = []fyne.CanvasObject{photoFinderView}
		case 3:
			contentContainer.Objects = []fyne.CanvasObject{cacheCleanerView}
		case 4:
			contentContainer.Objects = []fyne.CanvasObject{diskAnalyzerView}
		case 5:
			contentContainer.Objects = []fyne.CanvasObject{updaterView}
		}
		contentContainer.Refresh()
	}

	// Create sidebar
	sidebar := CreateSidebar()

	// Connect sidebar to navigation
	sidebarList := sidebar.(*container.Scroll).Content.(*widget.List)
	sidebarList.OnSelected = func(id widget.ListItemID) {
		onNavigate(id)
	}

	// Initialize with duplicate finder view
	onNavigate(0)

	// Main layout with split - sidebar on left, content on right
	split := container.NewHSplit(sidebar, contentContainer)
	split.Offset = 0.2 // Sidebar takes 20% of width

	// Main layout
	mainContent := container.NewBorder(
		header, // top
		nil,    // bottom
		nil,    // left (sidebar is in split)
		nil,    // right
		split,
	)

	return mainContent
}
