"""Analyze outlier posts with extreme total scores."""

import sqlite3
from pathlib import Path


def get_outlier_posts(cur: sqlite3.Cursor, model: str) -> dict:
    """Identify outlier posts for a model based on total score statistics."""
    
    # Calculate mean and std deviation
    query = """
        SELECT 
            reputation + power + wealth + achievement + pleasure + 
            independent_thoughts + independent_actions + stimulation + 
            personal_security + societal_security + tradition + lawfulness + 
            respect + humility + responsibility + caring + equality + nature + tolerance as total_score
        FROM analyses
        WHERE model = ?
    """
    cur.execute(query, (model,))
    scores = [row[0] for row in cur.fetchall() if row[0] is not None]
    
    if len(scores) < 3:
        return {"model": model, "count": 0}
    
    mean = sum(scores) / len(scores)
    variance = sum((s - mean) ** 2 for s in scores) / len(scores)
    std = variance ** 0.5
    
    # Define outliers (2 std deviations)
    outlier_high = mean + 2 * std
    outlier_low = max(0, mean - 2 * std)
    
    # Get outlier posts
    db_columns = [
        "reputation", "power", "wealth", "achievement", "pleasure",
        "independent_thoughts", "independent_actions", "stimulation",
        "personal_security", "societal_security", "tradition", "lawfulness",
        "respect", "humility", "responsibility", "caring", "equality", "nature", "tolerance"
    ]
    
    sum_expr = " + ".join(db_columns)
    
    query = f"""
        SELECT 
            a.post_at_uri,
            p.text,
            a.reasoning,
            {sum_expr} as total_score,
            a.reputation, a.power, a.wealth, a.achievement, a.pleasure,
            a.independent_thoughts, a.independent_actions, a.stimulation,
            a.personal_security, a.societal_security, a.tradition, a.lawfulness,
            a.respect, a.humility, a.responsibility, a.caring, a.equality, a.nature, a.tolerance
        FROM analyses a
        JOIN posts p ON a.post_at_uri = p.at_uri
        WHERE a.model = ?
        ORDER BY total_score DESC
        LIMIT 10
    """
    
    cur.execute(query, (model,))
    high_outliers = cur.fetchall()
    
    query = f"""
        SELECT 
            a.post_at_uri,
            p.text,
            a.reasoning,
            {sum_expr} as total_score,
            a.reputation, a.power, a.wealth, a.achievement, a.pleasure,
            a.independent_thoughts, a.independent_actions, a.stimulation,
            a.personal_security, a.societal_security, a.tradition, a.lawfulness,
            a.respect, a.humility, a.responsibility, a.caring, a.equality, a.nature, a.tolerance
        FROM analyses a
        JOIN posts p ON a.post_at_uri = p.at_uri
        WHERE a.model = ?
        ORDER BY total_score ASC
        LIMIT 10
    """
    
    cur.execute(query, (model,))
    low_outliers = cur.fetchall()
    
    return {
        "model": model,
        "count": len(scores),
        "mean": mean,
        "std": std,
        "outlier_threshold_high": outlier_high,
        "outlier_threshold_low": outlier_low,
        "high_outliers": high_outliers,
        "low_outliers": low_outliers,
    }


def format_outlier_report(outliers_data: dict) -> str:
    """Format outlier posts as human-readable report."""
    lines = []
    lines.append("=" * 80)
    lines.append(f"OUTLIER ANALYSIS FOR: {outliers_data['model']}")
    lines.append("=" * 80)
    lines.append(f"\nStatistics:")
    lines.append(f"  Total posts: {outliers_data['count']}")
    lines.append(f"  Mean total score: {outliers_data['mean']:.2f}")
    lines.append(f"  Std deviation: {outliers_data['std']:.2f}")
    lines.append(f"  High outlier threshold: {outliers_data['outlier_threshold_high']:.2f}")
    lines.append(f"  Low outlier threshold: {outliers_data['outlier_threshold_low']:.2f}")
    
    # Value names for formatting
    value_names = [
        "Reputation", "Power", "Wealth", "Achievement", "Pleasure",
        "Ind. Thoughts", "Ind. Actions", "Stimulation",
        "Personal Sec", "Societal Sec", "Tradition", "Lawfulness",
        "Respect", "Humility", "Responsibility", "Caring", "Equality", "Nature", "Tolerance"
    ]
    
    # High outliers
    lines.append("\n" + "=" * 80)
    lines.append("HIGH OUTLIERS (Top 10 posts by total score)")
    lines.append("=" * 80)
    
    for i, row in enumerate(outliers_data['high_outliers'], 1):
        uri, text, reasoning, total = row[0], row[1], row[2], row[3]
        values = row[4:]
        
        lines.append(f"\n[{i}] Total Score: {total}")
        lines.append(f"Post: {text[:150]}...")
        lines.append(f"Reasoning: {reasoning[:200]}...")
        lines.append(f"Values: {dict(zip(value_names, values))}")
    
    # Low outliers
    lines.append("\n" + "=" * 80)
    lines.append("LOW OUTLIERS (Bottom 10 posts by total score)")
    lines.append("=" * 80)
    
    for i, row in enumerate(outliers_data['low_outliers'], 1):
        uri, text, reasoning, total = row[0], row[1], row[2], row[3]
        values = row[4:]
        
        lines.append(f"\n[{i}] Total Score: {total}")
        lines.append(f"Post: {text[:150]}...")
        lines.append(f"Reasoning: {reasoning[:200]}...")
        lines.append(f"Values: {dict(zip(value_names, values))}")
    
    lines.append("\n" + "=" * 80)
    lines.append("END OF REPORT")
    lines.append("=" * 80)
    
    return "\n".join(lines)


def main():
    # Connect to database
    con = sqlite3.connect("../data.db")
    cur = con.cursor()
    
    # Models to analyze
    models = {
        "openai/gpt-4o-mini": "GPT-4o-mini",
        "mistralai/ministral-8b-2512": "Ministral-8b",
    }
    
    # Create output directory
    report_dir = Path("outliers")
    report_dir.mkdir(exist_ok=True)
    
    print("=" * 80)
    print("OUTLIER ANALYSIS")
    print("=" * 80)
    
    for model_name, display_name in models.items():
        print(f"\nAnalyzing {display_name}...")
        
        data = get_outlier_posts(cur, model_name)
        report = format_outlier_report(data)
        
        # Print to console
        print(report)
        
        # Save to file
        output_file = report_dir / f"{display_name.replace(' ', '_').lower()}_outliers.txt"
        with open(output_file, 'w', encoding='utf-8') as f:
            f.write(report)
        
        print(f"\nReport saved to: {output_file}")
    
    con.close()
    
    print("\n" + "=" * 80)
    print(f"DONE! Check the '{report_dir}/' directory for reports.")
    print("=" * 80)


if __name__ == "__main__":
    main()
