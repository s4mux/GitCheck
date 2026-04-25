# GitCheck – Engineering Principles

Diese Prinzipien gelten für alle Entscheidungen im Projekt – Architektur, Modulschnitt, Code. Sie sind keine Empfehlungen, sondern Leitlinien, gegen die wir aktiv prüfen.

---

## Single Responsibility Principle (SRP)

Jedes Package, jede Datei, jede Funktion hat genau eine Verantwortung.

**Konkret im Projekt:**
- `scanner` findet Repos – er bewertet sie nicht
- `git` befragt Repos – er formatiert nichts
- `report` formatiert – er führt keine Git-Operationen durch
- `config` lädt Konfiguration – er kennt keine Domänenlogik

**Warnsignal:** Ein Package importiert mehr als zwei andere interne Packages. Dann prüfen, ob die Verantwortung zu breit geworden ist.

---

## Dependency Inversion Principle (DIP)

High-level Module hängen nicht von Low-level Implementierungen ab – beide hängen von Abstraktionen ab.

**Konkret im Projekt:**
- `cmd` kennt keine konkreten Implementierungen – es verdrahtet
- Wenn `git` austauschbar sein soll (z.B. für Tests), wird ein Interface definiert, gegen das `cmd` arbeitet, nicht gegen das Package direkt

```go
// Kontrakt – nicht Implementierung
type RepoResolver interface {
    Build(path string, fetchRemote bool) (Repo, error)
}
```

**Praktische Grenze:** Interfaces entstehen, wenn Austauschbarkeit oder Testbarkeit es erfordern. Nicht präventiv für jede Funktion.

---

## High Cohesion, Low Coupling

Zusammengehöriges bleibt zusammen. Unabhängiges bleibt getrennt.

**Konkret im Projekt:**
- Alle Git-Operationen leben in `internal/git` – keine Git-Logik streut in andere Packages
- Ignore-Logik ist vollständig in `scanner/ignore.go` – `scanner.go` kennt keine Pattern-Matching-Details
- `RepoStatus` und `Repo` leben bei `git`, nicht bei `report` oder `cmd`

**Warnsignal:** Ein Package muss geändert werden, obwohl sich seine eigene Verantwortung nicht geändert hat. Dann ist die Kopplung zu hoch.

---

## Separation of Concerns

Fachliche Domäne und technische Aspekte werden getrennt gehalten.

**Konkret im Projekt:**
- Was ein Repo ist und welchen Status es hat → `internal/git` (Domäne)
- Wie Output formatiert und coloriert wird → `internal/report` (Technik)
- Wie das Filesystem traversiert wird → `internal/scanner` (Technik)
- Wie Konfiguration geladen wird → `internal/config` (Technik)

Domänentypen (`Repo`, `RepoStatus`) fließen durch die technischen Schichten – nicht umgekehrt.

---

## Modularisierung nur wenn gerechtfertigt

Ein neues Package oder eine neue Abstraktion entsteht, wenn mindestens eines zutrifft:

1. **Wiederverwendbarkeit steigt** – die Logik wird an mehr als einer Stelle gebraucht
2. **Austauschbarkeit steigt** – die Implementierung muss ersetzbar sein
3. **Komplexität wird lokal begrenzt** – ohne Trennung wächst die Komplexität ins Uferlose

**Explizit kein Grund:** Pattern um des Patterns willen. Kein Interface, keine Abstraktion, kein Package, das keinen dieser drei Punkte erfüllt.

---

## Keine defensive Überabstraktion

Wir schreiben Code für das Problem von heute, nicht für hypothetische Anforderungen von morgen.

- Keine generischen Frameworks für einen einzigen Use-Case
- Keine Plugin-Systeme, die niemand braucht
- Keine konfigurierbaren Abstraktionsebenen, wenn ein `if` reicht

Wenn morgen eine neue Anforderung kommt, refactorn wir dann – mit echtem Kontext, nicht mit Annahmen.

---

## Explizit über implizit

Verhalten soll lesbar sein, ohne dass man den Kontext kennen muss.

**Konkret:**
- Funktionssignaturen zeigen ihre Abhängigkeiten als Parameter – keine globalen Zustände
- Fehler werden explizit zurückgegeben und behandelt – kein `panic` außer bei echten Programmierfehlernh
- Konfiguration wird einmal geladen und weitergereicht – kein implizites Lesen aus Umgebungsvariablen tief im Code

```go
// Gut: explizit
func GetStatus(repoPath string, fetchRemote bool) (RepoStatus, error)

// Schlecht: implizit
func GetStatus(repoPath string) (RepoStatus, error) // liest intern irgendwo einen globalen Flag
```

---

## Testbarkeit ist kein Nachgedanke

Code wird so geschrieben, dass er testbar ist – ohne dass Testbarkeit die Architektur verbiegt.

- Funktionen nehmen `io.Writer` statt direkt auf `os.Stdout` zu schreiben
- Git-Operationen sind in einem Package isoliert, das gegen echte Temp-Repos getestet wird
- Kein Mocking des `git`-Binaries – Integration gegen echte Repos ist robuster

---

## Fehlerbehandlung

- Fehler werden so nah wie möglich am Ursprung mit Kontext angereichert
- `fmt.Errorf("scan %s: %w", path, err)` – nicht nackte Fehler weiterreichen
- Einzelne Repo-Fehler sind nicht fatal – sie werden gesammelt und am Ende gemeldet
- Konfigurationsfehler und fehlende Systemvoraussetzungen sind fatal (Exit 1)
