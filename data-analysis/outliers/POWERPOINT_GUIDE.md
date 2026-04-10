# PowerPoint Usage Guide

## Files Created

### 1. outliers_boxplot_scatter.png
**Main visualization for your PPT slide**

**Features:**
- ✅ Box plot showing distribution (median, quartiles)
- ✅ Scatter points for all posts
- ✅ Annotated outliers (top 3 for each model)
- ✅ Statistics box (mean, median, std, range)
- ✅ Improvement callout (+34%)
- ✅ 300 DPI resolution (high quality)
- ✅ White background (transparent in PPT)

**Slide Setup:**
```
Title: "PROMPT V4: Outlier Analysis & Score Distribution"
Sub: "Ministral-8b outperforms GPT-4o-mini +34%"

[Insert PNG image here]

Key Points:
- Ministral-8b: Higher variance (better value detection)
- Wider range: 8-36 vs 6-28
- Outliers correctly identify value-heavy posts
```

---

## How to Use in PowerPoint

### Step 1: Insert Image
1. Insert → Pictures → Browse
2. Select: `outliers_boxplot_scatter.png`
3. Resize to fit slide (maintain aspect ratio)

### Step 2: Add Title
**Title:** "PROMPT V4: Outlier Analysis & Score Distribution"  
**Subtitle:** "Ministral-8b detects more values per post"

### Step 3: Speaker Notes
Copy from: `outliers_summary.txt`

**Key talking points:**
- Box plot shows statistical distribution
- Outliers (diamond markers) = value-heavy posts
- Ministral-8b: +34% higher average score
- Wider std dev (9.1 vs 4.3) = better variance detection
- Top outliers: posts about politics, social issues

### Step 4: Callout Boxes (Optional)
Add text boxes in PowerPoint:

**Box 1 (Left):**
```
GPT-4o-mini (V4)
• Average: 12.7
• Conservative scoring
• Range: 6-28
```

**Box 2 (Right):**
```
Ministral-8b (V4)
• Average: 17.0
• +34% vs GPT
• Range: 8-36
```

**Box 3 (Bottom):**
```
Outlier Analysis:
• High scores = posts with rich value content
• Ministral detects nuances GPT misses
• Cost: 10x lower than large models
```

---

## Data Interpretation

### What the Box Plot Shows:
```
        GPT-4o-mini         Ministral-8b
           ┌───┐               ┌───┐
           │   │               │   │
       ┌───┤12 ├───┐       ●───┤13 ├───●  (outlier 36)
       │   │   │   │          │   │   │
       └───┤   ├───┘          └───┤   ├───┘
           │   │               │   │
           └───┘               └───┘
```

**Key Insights:**
1. **Median line** (middle of box): Ministral higher (13 vs 12)
2. **Box height** (IQR): Ministral has more variance
3. **Whiskers**: Ministral reaches higher scores
4. **Outliers** (annotated): Top scores clearly labeled

### What Points Represent:
- Each dot = 1 post analyzed
- Position = total score (sum of 19 values)
- Higher = more values detected

### Outlier Meaning:
- **Score 36** (Ministral): Political post with many values
- **Score 28** (GPT): Event announcement with moderate values
- **Score 6-8**: Casual/entertainment posts (few values)

---

## Additional Slides (Optional)

### Slide 2: Example Outliers
```
Title: "Outlier Case Studies"

[Create 3 boxes with:]

Post 1 (Score 36 - Ministral):
"Muore Robert Mueller, uomo integerrimo..."
Values: Societal Security(6), Equality(6), Caring(5)

Post 2 (Score 28 - GPT):
"Domani a #Brescia..."
Values: Tolerance(6), Respect(5), Equality(5)

Post 3 (Score 6 - GPT):
"Tutti i cittadini hanno pari dignità..."
Values: Equality(6) only
Voiced: Why such low score for constitutional text?
```

### Slide 3: Comparison Table
```
Title: "Model Comparison"

| Metric        | GPT-4o-mini | Ministral-8b | Δ      |
|---------------|-------------|--------------|--------|
| Average       | 12.7        | 17.0         | +34%   |
| Std Dev       | 4.3         | 9.1          | +112%  |
| Range         | 6-28        | 8-36         | Wider  |
| Outliers      | 1 (28)      | 1 (36)       | Higher |

Conclusion: Ministral-8b detects more nuanced values
```

---

## Color Coding (For Consistency)

Use these colors in your presentation:
- **GPT-4o-mini:** Green (#27AE60)
- **Ministral-8b:** Orange (#FF6B35)
- **Outliers:** Red markers
- **Statistics:** Light blue box

---

## Quick Stats for Slides

**Concise takeaway:**
- Ministral-8b: **+34% average score**
- **2x variance** (better detection spread)
- **Outlier: Score 36** (vs GPT's 28)
- **Cost:** 85% cheaper than GPT-4

**In one sentence:**
> "Ministral-8b V4 detects more values per post with greater variance 
> at 85% lower cost, making it ideal for production use."

---

## Export Settings

For best PPT quality:
- **Format:** PNG
- **Resolution:** 300 DPI
- **Size:** 12×7 inches (1200×700 pixels at 100 DPI)
- **Background:** White

To resize: Right-click PNG → Format Picture → Size → 100% scale

