package scanner

// Registry maps mode names to Scanner implementations
var scannerRegistry = map[string]func() Scanner{
	"audio": func() Scanner {
		return NewAudioScanner()
	},
	"byte": func() Scanner {
		return NewByteScanner()
	},
	"photo": func() Scanner {
		return NewPhotoScanner()
	},
}

// GetScanner returns a scanner instance for the given mode
func GetScanner(mode string) (Scanner, bool) {
	factory, ok := scannerRegistry[mode]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// AvailableModes returns a list of available scanner modes
func AvailableModes() []string {
	modes := make([]string, 0, len(scannerRegistry))
	for mode := range scannerRegistry {
		modes = append(modes, mode)
	}
	return modes
}
