# Russian Troll Tweets Analysis

## Overview

Analyze 3 million Russian troll tweets from the Internet Research Agency (IRA) using Schwartz Values framework to identify propaganda patterns and compare with normal social media content.

## Dataset Source

- **Origin**: FiveThirtyEight / Clemson University
- **Data**: ~3 million tweets from 2,848 Twitter handles
- **Period**: February 2012 - May 2018 (mostly 2015-2017)
- **Reference**: "Troll Factories: The Internet Research Agency and State-Sponsored Agenda Building"

## Research Questions

1. **Do troll tweets have distinct Schwartz value profiles?**
   - Higher Power/Achievement scores?
   - Lower Benevolence/Universalism?
   - More extreme scores (0 or 6)?

2. **Can Schwartz values identify manipulation/disinformation?**
   - Are there patterns that distinguish troll content from normal content?
   - Do different troll categories (RightTroll, LeftTroll) have different profiles?

3. **Distribution analysis**
   - Are troll scores more polarized (extremes)?
   - Do they cluster differently than normal posts?

## Hypotheses

- **Power**: Higher (hierarchy, control, manipulation)
- **Achievement**: Higher (success, ambition)
- **Conformity/Tradition**: Higher for RightTroll
- **Benevolence**: Lower (less empathy for others)
- **Universalism**: Lower (less concern for welfare of all)
- **Score distribution**: More extreme values (0 or 6) vs normal distribution

## Data Structure

### CSV Columns (relevant)
- `content` - Tweet text
- `author` - Handle name (e.g., "10_GOP")
- `publish_date` - When posted
- `language` - Filter for English
- `account_category` - Troll type (RightTroll, LeftTroll, etc.)
- `retweet` - Flag for retweets (0=original, 1=retweet)
- `followers`, `following` - Account metrics

### Account Categories
- RightTroll
- LeftTroll
- News
- HashtagGamer
- Commercial
- Fearmonger
- Unknown

## Implementation Plan

### Phase 1: Sample Analysis (Recommended)

**Budget-friendly approach:**
- Extract stratified sample: ~1000 tweets per category
- Filter: English only, original tweets (not retweets)
- Total: ~5000-7000 tweets
- Cost estimate: ~$5-10 in API credits

### Phase 2: Full Analysis

**If sample shows interesting patterns:**
- Analyze all original tweets (exclude retweets)
- Estimated: ~500k-1M unique original tweets
- Cost: ~$100-200 in API credits

### Phase 3: Comparison

Compare with existing Bluesky data:
- Match by language (English)
- Match by engagement metrics (likes, replies)
- Compare Schwartz value distributions
- Statistical significance tests

## Database Changes

### Schema Updates
```sql
ALTER TABLE posts ADD COLUMN source TEXT DEFAULT 'bluesky';
ALTER TABLE posts ADD COLUMN account_category TEXT;
ALTER TABLE posts ADD COLUMN external_id TEXT;
ALTER TABLE posts ADD COLUMN is_retweet INTEGER DEFAULT 0;
```

### Import Script
Create `feed-generator/import_troll_tweets.go`:
1. Read CSV files
2. Filter: English, non-retweet
3. Map columns to Post struct
4. Insert into database
5. Analyze with existing pipeline

## Expected Results

### Visualizations
1. **Value distribution heatmap** - Troll vs Normal
2. **Box plots** - Each Schwartz value comparison
3. **PCA/t-SNE** - Clustering of troll vs normal posts
4. **Radar charts** - Profile comparison by troll category
5. **Time series** - Value trends over time

### Statistical Analysis
- T-tests for each Schwartz value
- Effect sizes (Cohen's d)
- ROC curve for troll detection

## Alternative Approach: Separate Analysis

If budget is limited:
1. Don't import to DB
2. Analyze sample directly
3. Save results to CSV
4. Compare using Python/Pandas
5. Generate comparison visualizations

## File Locations

- Raw data: `IRAhandle_tweets_*.csv` (9 files)
- Import script: `feed-generator/import_troll.go`
- Analysis script: `data-analysis/troll_comparison.py`
- Results: `data-analysis/troll_analysis/`

## Next Steps

1. Decide sample size vs full dataset
2. Choose troll categories to analyze
3. Implement import script
4. Run analysis
5. Generate visualizations
6. Write findings

## Potential Findings

### Academic Contribution
- First Schwartz Values analysis of state-sponsored troll content
- Method for detecting manipulation via value analysis
- Comparison framework for different content sources

### Practical Applications
- Troll detection classifier
- Content moderation heuristics
- Platform trust frameworks

## References

- FiveThirtyEight: https://fivethirtyeight.com/features/why-were-sharing-3-million-russian-troll-tweets/
- Linvill & Warren paper: "Troll Factories: The Internet Research Agency and State-Sponsored Agenda Building"
- Schwartz Values: https://www.schwartzvalues.com/