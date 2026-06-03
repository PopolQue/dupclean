# Code-Review: DupClean Schwachstellen-Analyse

## KRITISCHE SCHWACHSTELLEN (Security & Safety)

### 1. Command Injection Risiko 🔴 KRITISCH [x]

    Betroffene Dateien: gui/app.go, cleaner/deleter.go, ui/ui.go

     1 // gui/app.go:748
     2 cmd = exec.Command("powershell", "-c",
     3     fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", path))

    Problem: Der path wird direkt in den Shell-Befehl eingefügt. Bei Pfaden mit speziellen Zeichen (', ", ;, |) könnte dies zu Command-Injection führen.

    Beispiel Angriff:

     1 Pfad: /music/test'; rm -rf /; '
     2 Ergebnis: (New-Object Media.SoundPlayer '/music/test'; rm -rf /; ').PlaySync()

    Betroffene Stellen:
     - gui/app.go:748 - Audio-Playback (PowerShell)
     - gui/app.go:696 - AppleScript (moveToTrashMacOS)
     - cleaner/deleter.go:229 - AppleScript
     - cleaner/deleter.go:278 - PowerShell Script

    ---

### 2. Race Conditions bei GUI-Operationen 🟠 HOCH [x]

    Betroffene Dateien: gui/app.go, gui/cache_cleaner.go

     1 // gui/app.go:318
     2 go func() {
     3     // ... Scan-Logik ...
     4     fyne.Do(func() {
     5         // UI-Update
     6     })
     7 }()

    Problem: Mehrere Goroutines greifen auf state.Groups, state.DeletedCount, state.FreedBytes zu ohne Synchronisation.

    Konkrete Issues:
     - state.Groups wird in keepAndDelete() modifiziert (Zeile 584) während UI darauf zugreift
     - state.CurrentGroupIndex wird ohne Mutex geändert
     - state.PlayingPath und state.CurrentPlayer haben Race-Bedingungen

    ---

### 3. Unsicherer Trash-Mechanismus 🟠 HOCH [x]

    Betroffene Dateien: cleaner/deleter.go, gui/app.go

      1 // cleaner/deleter.go:247-256
      2 trashName := filepath.Base(path)
      3 dest := filepath.Join(trashDir, trashName)
      4 counter := 1
      5 for {
      6     if _, err := os.Stat(dest); os.IsNotExist(err) {
      7         return os.Rename(path, dest)  // TOCTOU Race Condition!
      8     }
      9     // ...
     10 }

    Problem: Time-of-Check-to-Time-of-Use (TOCTOU) Race Condition. Zwischen os.Stat() und os.Rename() könnte eine andere Datei erstellt werden.

    ---

### 4. Symlink-Angriffsfläche 🟡 MITTEL [x]

    Betroffene Dateien: scanner/byte.go, scanner/photo.go, diskanalyzer/walker.go

     1 // scanner/byte.go:26
     2 err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
     3     // Keine Symlink-Prüfung!
     4     bySize[info.Size()] = append(bySize[info.Size()], path)

    Problem: Symlinks werden nicht konsistent geprüft. Ein Angreifer könnte Symlinks auf sensible Dateien platzieren, die dann gelesen/gelöscht werden könnten.

    Betroffene Stellen:
     - diskanalyzer/walker.go:137 prüft Symlinks, aber Scanner nicht
     - Inkonsistente Behandlung über Package hinweg

    ---

    ⚠️ HOHE PRIORITÄT (Stabilität & Robustheit)

### 5. Fehler werden geschluckt 🟠 HOCH [x]

    Betroffene Dateien: scanner/byte.go, scanner/photo.go, cleaner/scanner.go

      1 // scanner/byte.go:27
      2 if err != nil {
      3     return nil // skip unreadable files  ❌ Kein Logging!
      4 }
      5
      6 // scanner/photo.go:106
      7 if err != nil {
      8     return nil, stats, err  // ❌ Kontext fehlt
      9 }
     10
     11 // cleaner/scanner.go:117
     12 if _, err := os.Stat(path); err != nil {
     13     continue // ❌ Still schweigend überspringen
     14 }

    Problem: Fehler werden ohne Kontext oder Logging ignoriert. Production-Software muss protokollieren, was schiefgeht.

    ---

### 6. Goroutine-Leaks 🟠 HOCH [x]

    Betroffene Dateien: gui/app.go, gui/cache_cleaner.go, cleaner/deleter.go

     1 // gui/app.go:765
     2 go func() {
     3     _ = cmd.Run()
     4     if state.CurrentPlayer == cmd {
     5         state.CurrentPlayer = nil  // ❌ Kein Lock!
     6         state.StopPlayer = nil
     7         state.PlayingPath = ""
     8     }
     9 }()

    Problem:
     - Goroutines werden gestartet, aber bei GUI-Navigation nicht sauber beendet
     - stopPlayback() wird aufgerufen, aber Race-Bedingungen bleiben

    ---

### 7. Pfad-Validierung unzureichend 🟠 HOCH [x]

    Betroffene Dateien: cleaner/deleter.go, gui/cache_cleaner.go

     1 // cleaner/deleter.go:133-137
     2 if entry.Path == "" {
     3     return 0, 0, false, fmt.Errorf("cannot delete empty path")
     4 }
     5 if entry.Path == "/" || entry.Path == `\` {
     6     return 0, 0, false, fmt.Errorf("cannot delete root directory")
     7 }

    Problem: Diese Checks sind unzureichend. Was ist mit:
     - /home/user/../../../ (Path Traversal)
     - C:\Windows\System32 (Windows System-Ordner)
     - /etc, /bin, /usr (Linux System-Ordner)

    ---

### 8. Memory-Belastung bei großen Scans 🟡 MITTEL [x]

    Betroffene Dateien: scanner/byte.go, diskanalyzer/walker.go

     1 // scanner/byte.go:24
     2 bySize := make(map[int64][]string)  // ❌ Alle Pfade im Speicher!
     3
     4 // diskanalyzer/walker.go:95
     5 func statPass(root string, opts WalkOptions) ([]FileEntry, []error, error) {
     6     var entries []FileEntry  // ❌ Alle Entries im Speicher!

    Problem: Bei großen Verzeichnissen (100k+ Dateien) wird sehr viel Speicher benötigt. Keine Streaming-Verarbeitung.

    ---

## 📋 MITTLERE PRIORITÄT (Code-Qualität)

### 9. Duplizierter Code 🟡 MITTEL [x]

    Betroffene Dateien: gui/app.go, cleaner/deleter.go, ui/ui.go

     1 // Drei fast identische moveToTrash-Implementierungen:
     2 // - gui/app.go:683-718
     3 // - cleaner/deleter.go:213-278
     4 // - ui/ui.go:132-150

    Problem: Code-Duplizierung führt zu:
     - Inkonsistentem Verhalten
     - Mehrfachem Fixen desselben Bugs
     - Höherem Wartungsaufwand

    ---

### 10. Fehlende Input-Validierung 🟡 MITTEL [x]

    Betroffene Dateien: gui/app.go

     1 // gui/app.go:780-795
     2 func showIgnoreDialog(state *AppState, onConfirm func()) {
     3     extensionsEntry := widget.NewEntry()
     4     extensionsEntry.SetPlaceHolder("e.g. .txt, .pdf, .jpg")
     5     // ...
     6     state.IgnoreExtensions = append(state.IgnoreExtensions, strings.ToLower(ext))

    Problem: Keine Validierung der Extensions. Regex-Injection möglich bei .txt.* oder *.

    ---

### 11. Hardcoded Pfade 🟡 MITTEL [x]

    Betroffene Dateien: gui/app.go

     1 // gui/app.go:28
     2 logFile, err := os.OpenFile("/tmp/dupclean.log", ...)

    Problem:
     - /tmp existiert nicht auf Windows
     - Keine Konfigurationsmöglichkeit
     - Berechtigungsprobleme bei eingeschränkten Usern

    ---

### 12. Inkonsistente Error-Types 🟡 MITTEL [x]

    Betroffene Dateien: cleaner/deleter.go, scanner/types.go

     1 // cleaner/deleter.go:28
     2 type DeleteError struct {
     3     Path    string
     4     Err     error
     5     Skipped bool
     6 }
     7
     8 // scanner - keine custom Error-Types!

    Problem: Jeder Package-Teil hat eigene Error-Behandlung. Keine einheitliche Fehlerklassifikation.

    ---

## 📝 GERINGE PRIORITÄT (Best Practices)

### 13. TODO-Kommentar nicht umgesetzt [x]

    Betroffene Dateien: diskanalyzer/render_cli.go:214

     1 // TODO: Implement using golang.org/x/term
     2 func GetTerminalWidth(defaultWidth int) int {
     3     return defaultWidth  // ❌ TODO seit unbestimmter Zeit
     4 }

    ---

### 14. Magic Numbers [x]

    Betroffene Dateien: scanner/utils.go, scanner/byte.go

     1 // scanner/utils.go:11
     2 const partialHashSize = 8 * 1024 // 8KB
     3
     4 // scanner/byte.go:119
     5 const bufSize = 32 * 1024 // 32KB buffers

    Problem: Magic Numbers sollten benannt und dokumentiert sein. Warum 8KB? Warum 32KB?

    ---

### 15. Fehlende Context-Propagation [x]

    Betroffene Dateien: Alle Scanner, cleaner/scanner.go

     1 // Keine einzige Funktion akzeptiert context.Context!
     2 func (s *ByteScanner) Scan(root string, opts Options) (Result, error)

    Problem: Lange Operationen können nicht abgebrochen werden.

    ---

    📊 Zusammenfassung nach Kategorie


    ┌─────────────────────┬────────┬─────────────────────────┐
    │ Kategorie           │ Anzahl │ Schweregrad             │
    ├─────────────────────┼────────┼─────────────────────────┤
    │ Security            │ 4      │ 🔴 Kritisch - 🟡 Mittel │
    │ Race Conditions     │ 3      │ 🟠 Hoch                 │
    │ Error-Handling      │ 2      │ 🟠 Hoch - 🟡 Mittel     │
    │ Resource-Management │ 2      │ 🟠 Hoch - 🟡 Mittel     │
    │ Code-Qualität       │ 4      │ 🟡 Mittel               │
    └─────────────────────┴────────┴─────────────────────────┘

    ---

## 🎯 Empfohlene Priorisierung

    Sofort beheben (P0):
     1. Command-Injection verhindern (String-Escaping für Shell-Befehle)
     2. Race Conditions in GUI mit Mutex schützen
     3. TOCTOU-Race bei Trash-Operation fixen

    Kurzfristig (P1):
     4. Symlink-Validierung konsistent implementieren
     5. Error-Logging hinzufügen (nicht schlucken!)
     6. Goroutine-Leaks beheben
     7. Pfad-Validierung verbessern

    Mittelfristig (P2):
     8. Code-Duplizierung entfernen (DRY)
     9. Context-Propagation einführen
     10. Memory-Optimierung für große Scans

    ---

    Gesamteinschätzung: Das Projekt ist funktional, hat aber signifikante Security- und Stabilitätsprobleme, die vor einem Production-Einsatz unbedingt behoben werden müssen. Die größte
     Sorge ist die Kombination aus Command-Injection-Risiko und unsauberer Goroutine-Synchronisation.
