# Schwartz Values Feed Generator

Analizza post Bluesky con i valori Schwartz tramite AI.

## Flusso

```
Post URL → Fetch dati Bluesky → Build prompt → AI Provider → JSON output
```

1. Fetch post da Bluesky (testo, link, immagini, metadata)
2. Costruisce prompt con dati post
3. Chiama AI provider (OpenRouter / SiliconFlow) per analisi valori Schwartz
4. Salva risultati nel database SQLite

## Configurazione

Creare `.env`:

```bash
BSKY_HANDLE=tuo_handle.bsky.social
BSKY_APP_PASSWORD=tua_app_password
OPEN_ROUTER_KEY=tua_openrouter_key
SILICONFLOW_API_KEY=tua_siliconflow_key
```

## Comandi

### Build
```bash
cd feed-generator
make build
```

### Collect (default - raccoglie post da Bluesky)
```bash
./bin/feedgen                    # raccoglie 40 post (10 per cluster)
./bin/feedgen -n 100             # raccoglie 100 post
./bin/feedgen -lang it           # filtra per lingua (it, en, es, etc.)
```

### Analyze (analizza i post con l'AI)
```bash
./bin/feedgen analyze                       # analizza tutti i post non analizzati
./bin/feedgen -l 10 analyze                 # limita a 10 post
./bin/feedgen -model=qwen analyze            # filtra per modello (NOTA: flag prima di analyze)
./bin/feedgen -images=false analyze          # analisi solo testo (senza immagini)
./bin/feedgen -model=qwen -l 5 analyze       # combina filtri
```

**NOTA IMPORTANTE**: I flag devono essere posizionati PRIMA del comando `analyze`, non dopo.

### From File (carica post da file JSON)
```bash
./bin/feedgen from-file <urls.json>         # carica post da file JSON con URL Bluesky
```

## Modelli AI configurati

| Nome | Provider | Origine |
|------|----------|---------|
| gpt-4o-mini | OpenRouter | USA |
| ministral-8b | OpenRouter | EU |
| qwen3-vl-8b | SiliconFlow | CN |

Limite: massimo 5 analisi per post per ogni modello.

## Flag disponibili

| Flag | Default | Descrizione |
|------|---------|-------------|
| `-n` | 40 | Numero totale post da raccogliere |
| `-lang` | it | Lingua dei post (it, en, es, etc.) |
| `-l` | 0 | Limite post da analizzare (0 = tutti) |
| `-model` | "" | Filtro modello (match parziale) |
| `-images` | true | Includi immagini nell'analisi |
| `-limit` | 0 | Numero post per modello |