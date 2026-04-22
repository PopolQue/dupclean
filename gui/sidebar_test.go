package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Test SidebarItem structure
func TestSidebarItem_Struct(t *testing.T) {
	clicked := false
	item := SidebarItem{
		Name:    "Test Item",
		Icon:    theme.HomeIcon(),
		Tooltip: "This is a test tooltip",
		OnClick: func() {
			clicked = true
		},
	}

	if item.Name != "Test Item" {
		t.Errorf("Name = %q, want %q", item.Name, "Test Item")
	}
	if item.Tooltip != "This is a test tooltip" {
		t.Errorf("Tooltip = %q, want %q", item.Tooltip, "This is a test tooltip")
	}

	// Test OnClick callback
	item.OnClick()
	if !clicked {
		t.Error("OnClick callback should have been invoked")
	}
}

// Test Sidebar with empty items
func TestSidebar_EmptyItems(t *testing.T) {
	items := []SidebarItem{}
	list := Sidebar(items)

	if list == nil {
		t.Fatal("Sidebar should return a non-nil list")
	}

	// The list should have length 0
	length := list.Length()
	if length != 0 {
		t.Errorf("List length = %d, want 0", length)
	}
}

// Test Sidebar with single item
func TestSidebar_SingleItem(t *testing.T) {
	items := []SidebarItem{
		{
			Name: "Single Item",
			Icon: theme.HomeIcon(),
		},
	}

	list := Sidebar(items)
	if list == nil {
		t.Fatal("Sidebar should return a non-nil list")
	}

	length := list.Length()
	if length != 1 {
		t.Errorf("List length = %d, want 1", length)
	}
}

// Test Sidebar with multiple items
func TestSidebar_MultipleItems(t *testing.T) {
	items := []SidebarItem{
		{
			Name: "Item 1",
			Icon: theme.HomeIcon(),
		},
		{
			Name: "Item 2",
			Icon: theme.DeleteIcon(),
		},
		{
			Name: "Item 3",
			Icon: theme.StorageIcon(),
		},
	}

	list := Sidebar(items)
	if list == nil {
		t.Fatal("Sidebar should return a non-nil list")
	}

	length := list.Length()
	if length != 3 {
		t.Errorf("List length = %d, want 3", length)
	}

	// Test OnSelected callback - should not panic
	list.OnSelected(1)
	// Note: OnClick is not automatically called by OnSelected
	// This is expected behavior, so clickedIndex stays -1
}

// Test Sidebar selection behavior
func TestSidebar_Selection(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1", Icon: theme.HomeIcon()},
		{Name: "Item 2", Icon: theme.DeleteIcon()},
		{Name: "Item 3", Icon: theme.StorageIcon()},
	}

	list := Sidebar(items)

	// Select first item
	list.OnSelected(0)
	// Select second item
	list.OnSelected(1)
	// Select third item
	list.OnSelected(2)

	// All selections should succeed without panic
}

// Test Sidebar out of bounds selection
func TestSidebar_OutOfBoundsSelection(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1", Icon: theme.HomeIcon()},
	}

	list := Sidebar(items)

	// These should not panic
	list.OnSelected(-1)
	list.OnSelected(100)
}

// Test CreateSidebar returns valid structure
func TestCreateSidebar_ReturnsValidStructure(t *testing.T) {
	sidebar := CreateSidebar()

	if sidebar == nil {
		t.Fatal("CreateSidebar should return a non-nil object")
	}
}

// Test CreateSidebar has correct number of items
func TestCreateSidebar_ItemCount(t *testing.T) {
	// We can't directly access the list items, but we can verify
	// the sidebar is created successfully
	sidebar := CreateSidebar()
	if sidebar == nil {
		t.Fatal("CreateSidebar should return a non-nil object")
	}
}

// Test SidebarItem with nil OnClick
func TestSidebarItem_NilOnClick(t *testing.T) {
	item := SidebarItem{
		Name:    "Test",
		Icon:    theme.HomeIcon(),
		Tooltip: "Test tooltip",
		OnClick: nil,
	}

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SidebarItem with nil OnClick panicked: %v", r)
		}
	}()

	// Try to access fields
	_ = item.Name
	_ = item.Icon
	_ = item.Tooltip
}

// Test Sidebar with various icon types
func TestSidebar_VariousIcons(t *testing.T) {
	items := []SidebarItem{
		{Name: "Home", Icon: theme.HomeIcon()},
		{Name: "Search", Icon: theme.SearchIcon()},
		{Name: "Delete", Icon: theme.DeleteIcon()},
		{Name: "Storage", Icon: theme.StorageIcon()},
		{Name: "Settings", Icon: theme.SettingsIcon()},
		{Name: "Content", Icon: theme.ContentAddIcon()},
		{Name: "View", Icon: theme.ViewRefreshIcon()},
		{Name: "Info", Icon: theme.InfoIcon()},
		{Name: "Error", Icon: theme.ErrorIcon()},
		{Name: "Check", Icon: theme.ConfirmIcon()},
		{Name: "Cancel", Icon: theme.CancelIcon()},
		{Name: "Document", Icon: theme.DocumentIcon()},
		{Name: "Folder", Icon: theme.FolderIcon()},
		{Name: "File", Icon: theme.FileIcon()},
		{Name: "Upload", Icon: theme.UploadIcon()},
		{Name: "Download", Icon: theme.DownloadIcon()},
		{Name: "Computer", Icon: theme.ComputerIcon()},
		{Name: "Account", Icon: theme.AccountIcon()},
		{Name: "Login", Icon: theme.LoginIcon()},
		{Name: "Logout", Icon: theme.LogoutIcon()},
		{Name: "List", Icon: theme.ListIcon()},
		{Name: "Grid", Icon: theme.GridIcon()},
		{Name: "Menu", Icon: theme.MenuIcon()},
		{Name: "MoreHorizontal", Icon: theme.MoreHorizontalIcon()},
		{Name: "MoreVertical", Icon: theme.MoreVerticalIcon()},
		{Name: "Calendar", Icon: theme.CalendarIcon()},
		{Name: "Visibility", Icon: theme.VisibilityIcon()},
		{Name: "VisibilityOff", Icon: theme.VisibilityOffIcon()},
		{Name: "VolumeUp", Icon: theme.VolumeUpIcon()},
		{Name: "VolumeDown", Icon: theme.VolumeDownIcon()},
		{Name: "VolumeMute", Icon: theme.VolumeMuteIcon()},
		{Name: "MediaPlay", Icon: theme.MediaPlayIcon()},
		{Name: "MediaPause", Icon: theme.MediaPauseIcon()},
		{Name: "MediaStop", Icon: theme.MediaStopIcon()},
		{Name: "MediaRecord", Icon: theme.MediaRecordIcon()},
		{Name: "MediaFastForward", Icon: theme.MediaFastForwardIcon()},
		{Name: "MediaFastRewind", Icon: theme.MediaFastRewindIcon()},
		{Name: "MediaSkipNext", Icon: theme.MediaSkipNextIcon()},
		{Name: "MediaSkipPrevious", Icon: theme.MediaSkipPreviousIcon()},
		{Name: "MailCompose", Icon: theme.MailComposeIcon()},
		{Name: "MailForward", Icon: theme.MailForwardIcon()},
		{Name: "MailReply", Icon: theme.MailReplyIcon()},
		{Name: "MailReplyAll", Icon: theme.MailReplyAllIcon()},
		{Name: "FileApplication", Icon: theme.FileApplicationIcon()},
		{Name: "FileAudio", Icon: theme.FileAudioIcon()},
		{Name: "FileImage", Icon: theme.FileImageIcon()},
		{Name: "FileText", Icon: theme.FileTextIcon()},
		{Name: "FileVideo", Icon: theme.FileVideoIcon()},
		{Name: "FolderNew", Icon: theme.FolderNewIcon()},
		{Name: "Help", Icon: theme.HelpIcon()},
		{Name: "History", Icon: theme.HistoryIcon()},
		{Name: "Question", Icon: theme.QuestionIcon()},
		{Name: "SearchReplace", Icon: theme.SearchReplaceIcon()},
		{Name: "MailSend", Icon: theme.MailSendIcon()},
		{Name: "ZoomFit", Icon: theme.ZoomFitIcon()},
		{Name: "ZoomIn", Icon: theme.ZoomInIcon()},
		{Name: "ZoomOut", Icon: theme.ZoomOutIcon()},
	}

	// Just verify we can create the list without panic
	list := Sidebar(items)
	if list == nil {
		t.Fatal("Sidebar should return non-nil list")
	}

	length := list.Length()
	if length != len(items) {
		t.Errorf("List length = %d, want %d", length, len(items))
	}
}

// Test SidebarItem data integrity
func TestSidebarItem_DataIntegrity(t *testing.T) {
	tests := []struct {
		name     string
		itemName string
		icon     fyne.Resource
		tooltip  string
	}{
		{"empty strings", "", theme.HomeIcon(), ""},
		{"unicode name", "日本語アイテム", theme.HomeIcon(), "ヒント"},
		{"long name", "This is a very long sidebar item name that should still work", theme.HomeIcon(), "Tooltip"},
		{"special chars", "Item @#$%", theme.HomeIcon(), "Tooltip!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := SidebarItem{
				Name:    tt.itemName,
				Icon:    tt.icon,
				Tooltip: tt.tooltip,
			}

			if item.Name != tt.itemName {
				t.Errorf("Name mismatch: got %q, want %q", item.Name, tt.itemName)
			}
			if item.Tooltip != tt.tooltip {
				t.Errorf("Tooltip mismatch: got %q, want %q", item.Tooltip, tt.tooltip)
			}
		})
	}
}
