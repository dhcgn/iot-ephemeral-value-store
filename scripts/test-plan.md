# IoT Ephemeral Value Store - MCP Testplan

Testplan zur Validierung aller Operationen des IoT Ephemeral Value Store MCP-Servers.
Kann manuell oder durch einen KI-Agenten (z.B. Claude Code, Github Copilot CLI) wiederholt durchgeführt werden.

## Voraussetzungen

- MCP-Server `IoT-Ephemeral-Value-Store-Server` ist konfiguriert und erreichbar
- Zugriff auf alle 5 MCP-Tools:
  - `generate_key_pair`
  - `upload_data`
  - `patch_data`
  - `download_data`
  - `delete_data`

## Testvariablen

Die folgenden Werte werden in Test 1 erzeugt und in allen weiteren Tests verwendet:

| Variable         | Beschreibung                          |
|------------------|---------------------------------------|
| `UPLOAD_KEY`     | Wird in Test 1 generiert              |
| `DOWNLOAD_KEY`   | Wird in Test 1 generiert              |

---

## Test 1: Schlüsselpaar generieren

**Tool:** `generate_key_pair`

**Aktion:**
Rufe `generate_key_pair` ohne Parameter auf.

**Erwartetes Ergebnis:**
- Antwort enthält `upload_key` (Format: `u_` + 64 Hex-Zeichen)
- Antwort enthält `download_key` (Format: `d_` + 64 Hex-Zeichen)
- Erfolgsmeldung vorhanden

**Nachbereitung:** Speichere beide Schlüssel als `UPLOAD_KEY` und `DOWNLOAD_KEY` für alle weiteren Tests.

---

## Test 2: Daten hochladen (upload_data)

**Tool:** `upload_data`

**Aktion:**
```json
{
  "upload_key": "UPLOAD_KEY",
  "parameters": {
    "temperature": "22.5",
    "humidity": "45",
    "device": "sensor_livingroom"
  }
}
```

**Erwartetes Ergebnis:**
- Erfolgsmeldung `"Data uploaded successfully"`
- `parameter_count` = 3

---

## Test 3: Alle Daten herunterladen (download_data)

**Tool:** `download_data`

**Aktion:**
```json
{
  "download_key": "DOWNLOAD_KEY"
}
```

**Erwartetes Ergebnis:**
- `data` enthält alle 3 hochgeladenen Parameter:
  - `temperature` = `"22.5"`
  - `humidity` = `"45"`
  - `device` = `"sensor_livingroom"`
- Ein automatischer `timestamp` im ISO-8601-Format ist vorhanden (z.B. `"2026-02-22T17:04:17Z"`)

---

## Test 4: Einzelnen Parameter abfragen

**Tool:** `download_data`

**Aktion:**
```json
{
  "download_key": "DOWNLOAD_KEY",
  "parameter": "temperature"
}
```

**Erwartetes Ergebnis:**
- `data` = `"22.5"`
- `parameter` = `"temperature"`

---

## Test 5: Daten mergen (patch_data)

**Tool:** `patch_data`

**Aktion:**
```json
{
  "upload_key": "UPLOAD_KEY",
  "path": "",
  "parameters": {
    "pressure": "1013",
    "light_level": "750"
  }
}
```

**Erwartetes Ergebnis (nach Merge):**
- Erfolgsmeldung `"Data merged successfully"`
- `parameter_count` = 2

**Validierung:** Lade danach alle Daten herunter (`download_data` ohne `parameter`).

**Erwartetes Ergebnis der Validierung:**
- Alte Werte sind erhalten: `temperature`, `humidity`, `device`
- Neue Werte sind vorhanden: `pressure` = `"1013"`, `light_level` = `"750"`
- `timestamp` wurde aktualisiert

---

## Test 6: Verschachtelte Pfade (patch_data mit Pfad)

**Tool:** `patch_data`

**Aktion:**
```json
{
  "upload_key": "UPLOAD_KEY",
  "path": "bedroom/sensors",
  "parameters": {
    "temperature": "19.8",
    "humidity": "52"
  }
}
```

**Erwartetes Ergebnis:**
- Erfolgsmeldung mit `path` = `"bedroom/sensors"`

**Validierung A:** Subobjekt abrufen mit `download_data`:
```json
{
  "download_key": "DOWNLOAD_KEY",
  "parameter": "bedroom/sensors"
}
```
- `data` = `{ "humidity": "52", "temperature": "19.8" }`

**Validierung B:** Einzelwert im verschachtelten Pfad abrufen:
```json
{
  "download_key": "DOWNLOAD_KEY",
  "parameter": "bedroom/sensors/temperature"
}
```
- `data` = `"19.8"`

---

## Test 7: Daten löschen (delete_data)

**Tool:** `delete_data`

**Aktion:**
```json
{
  "upload_key": "UPLOAD_KEY"
}
```

**Erwartetes Ergebnis:**
- `success` = `true`
- Erfolgsmeldung `"Data deleted successfully"`

---

## Test 8: Download nach Löschung (Negativtest)

**Tool:** `download_data`

**Aktion:**
```json
{
  "download_key": "DOWNLOAD_KEY"
}
```

**Erwartetes Ergebnis:**
- Fehler / Error-Antwort
- Meldung enthält `"Key not found"` oder vergleichbare Fehlermeldung
- Keine Daten werden zurückgegeben

---

## Zusammenfassung / Ergebnisprotokoll

| # | Testname                              | Ergebnis | Anmerkung |
|---|---------------------------------------|----------|-----------|
| 1 | Schlüsselpaar generieren              |          |           |
| 2 | Daten hochladen (upload_data)         |          |           |
| 3 | Alle Daten herunterladen              |          |           |
| 4 | Einzelnen Parameter abfragen          |          |           |
| 5 | Daten mergen (patch_data)             |          |           |
| 6 | Verschachtelte Pfade (nested path)    |          |           |
| 7 | Daten löschen (delete_data)           |          |           |
| 8 | Download nach Löschung (Negativtest)  |          |           |

**Gesamtergebnis:** ___ / 8 bestanden

---

## Letzter erfolgreicher Durchlauf

- **Datum:** 2026-02-22
- **Ergebnis:** 8/8 bestanden
- **Bemerkungen:** Alle Operationen (CRUD + Merge + Nested Paths) funktionieren wie erwartet. Timestamps werden automatisch gesetzt und bei Updates aktualisiert.
