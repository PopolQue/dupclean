# Plan: Production-Grade-Upgrades für DupClean

## Phase 1: Kritische Infrastruktur (P0) [ ]

### 1. Graceful Shutdown mit Context [ ]

- [ ] Context-Propagation durch alle Scanner
- [ ] SIGINT/SIGTERM-Handling in main.go
- [ ] Sauberes Beenden von Goroutines

### 2. Strukturiertes Logging [ ]

- [ ] Logger-Interface mit Levels (Debug, Info, Warn, Error)
- [ ] JSON-Output für Production
- [ ] Korrelations-IDs für Request-Tracing

### 3. Error-Handling mit Kontext [ ]

- [ ] github.com/pkg/errors für Wrap/WithStack
- [ ] Custom Error-Types für verschiedene Fehlerklassen
- [ ] Konsistente Fehlerbehandlung (nicht schlucken!)

## Phase 2: Sicherheit & Konfiguration (P1) [ ]

### 4. Konfigurationssystem [ ]

- [ ] Config-Struct mit YAML/JSON-Support
- [ ] Environment-Variable-Integration
- [ ] CLI-Flags haben Vorrang vor Config

### 5. Sicherheit & Validierung [ ]

- [ ] Pfad-Validierung (Path-Traversal-Prevention)
- [ ] Symlink-Erkennung und -Handling
- [ ] Input-Validierung für alle Benutzer-Eingaben

### 6. Test-Coverage verbessern [ ]

- [ ] Coverage von 44% auf ≥80% für Kern-Logik
- [ ] Interface-basierte Mocks für externe Abhängigkeiten
- [ ] Konsistente Table-driven Tests

## Phase 3: Dokumentation & Monitoring (P2) [ ]

### 7. Godoc-Dokumentation [ ]

- [ ] Alle öffentlichen Funktionen/Types dokumentieren
- [ ] Beispiele für komplexe Funktionen

### 8. Performance-Monitoring [ ]

- [ ] Metriken für Scan-Dauer, Dateien, Duplikate
- [ ] Profiling-Support (pprof)

### 9. CI/CD-Erweiterungen [ ]

- [ ] Security-Scans (gosec, Trivy)
- [ ] Benchmark-Tests mit Regression-Checking Gewünschte Priorisierung?
