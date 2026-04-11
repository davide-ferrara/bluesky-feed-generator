# Perceptron Scoring Analysis

## Current Implementation

Il sistema di scoring attuale è un **single-layer perceptron**:

```
score = Σ(value_i × weight_i)
```

| Componente | Equivalente Perceptron |
|-------------|----------------------|
| 19 value scores (0-6) | Input x |
| User weights (0-10) | Weights w |
| Weighted sum | w · x (dot product) |
| Score finale | Output o |

## Problemi Identificati

1. **No bounds**: score può essere qualsiasi numero
2. **Non comparabile**: score 50 per utente A ≠ score 50 per utente B  
3. **Non leggibile**: non sai se score alto = "molto rilevante" o "meh"

## Soluzioni Proposte

### 1. Normalizzazione Input

Normalizzare i value scores prima della weighted sum:

```
normalized = (value_score × weight) / (6 × total_weights_sum)
```

Output finale: **0-100%** comparabile.

### 2. Activation Functions

| Function | Formula | Output | Note |
|----------|---------|--------|-------|
| None (current) | Σ(w·x) | any | Già funziona con pesi positivi |
| Sigmoid | 1/(1+e^-x) | (0,1) | Satura presto |
| Tanh | tanh(x) | (-1,1) | Permette pesi negativi |

### 3. Negative Weights (Likes + Dislikes)

Permettere pesi negativi per indicare "non voglio":
- Weight +5: "mi piace questo valore"
- Weight -5: "non mi piace questo valore"

Richiede tanh come activation.

## Esempi

### Senza Normalizzazione

| Utente | Pesi Totali | Score |
|-------|------------|-------|
| Alice | 20 | 66 (100%) |
| Bob | 20 | 22 (100%) |

Entrambi satura a 100% - non distinguibili.

### Con Normalizzazione

```
score_normalized = raw_score / (6 × sum_weights)
```

| Utente | Raw Score | Normalized |
|--------|-----------|------------|
| Alice | 66 | 55% |
| Bob | 22 | 18% |
| Charlie | 40 | 33% |

Ora comparabili: 55% vs 18% vs 33%.

## Todo

- [ ] Discutere col professore approccio preferito
- [ ] Implementare normalizzazione se approvato
- [ ] Considerare negative weights (likes + dislikes)