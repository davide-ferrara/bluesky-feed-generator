"""Main script for analyzing model performance from SQL database."""

import sqlite3
from pathlib import Path

from db_stats import (
    calculate_values_avg,
    calculate_cost_stats,
    calculate_time_stats,
    calculate_score_distribution,
    calculate_total_score_distribution,
)
from export import save_to_csv, generate_plots
from charts import plot_score_distribution_heatmap, plot_total_score_histogram


def main():
    # Connect to database
    con = sqlite3.connect("../data.db")
    cur = con.cursor()

    # Create plot directory
    plot_dir = Path("plot")
    plot_dir.mkdir(exist_ok=True)

    # Model name mapping (technical name -> display name)
    # To change models, modify this dictionary
    models = {
        "openai/gpt-4o-mini": "GPT-4o-mini",
        "mistralai/ministral-14b-2512": "Ministral-14b",
        "Qwen/Qwen3-VL-8B-Instruct": "Qwen-3",
        # "anthropic/claude-sonnet-4.6": "Claude 4",
        # "openai/gpt-4.1": "GPT-4.1",
        # Add more models here as needed
    }

    # Calculate statistics for each model
    stats = {}
    distributions = {}
    total_scores = {}
    for model_name, display_name in models.items():
        stats[display_name] = {
            "avg_values": calculate_values_avg(cur, model_name),
            "cost_stats": calculate_cost_stats(cur, model_name),
            "time_stats": calculate_time_stats(cur, model_name),
        }
        distributions[display_name] = calculate_score_distribution(cur, model_name)
        total_scores[display_name] = calculate_total_score_distribution(cur, model_name)

    # Print statistics
    print("=" * 70)
    print("STATISTICS SUMMARY")
    print("=" * 70)
    for display_name, data in stats.items():
        print(f"\n{display_name}:")
        print(f"  Values average: {data['avg_values']}")
        print(
            f"  Cost: avg=${data['cost_stats']['avg_cost']:.6f}, "
            f"total=${data['cost_stats']['total_cost']:.4f}, "
            f"posts={data['cost_stats']['num_posts']}"
        )
        print(
            f"  Time: avg={data['time_stats']['avg_time_ms']:.0f}ms, "
            f"total={data['time_stats']['total_time_ms'] / 1000:.1f}s"
        )

    # Generate plots and CSV
    print("\n" + "=" * 70)
    print("GENERATING OUTPUTS")
    print("=" * 70)
    model_names = list(models.values())

    generate_plots(stats, model_names, plot_dir)
    save_to_csv(stats, plot_dir)

    # Save distribution CSV
    from export import save_distribution_csv

    save_distribution_csv(distributions, plot_dir)

    # Generate score distribution heatmap
    print("\nGenerating score distribution heatmap...")
    plot_score_distribution_heatmap(
        distributions=distributions,
        output_path=str(plot_dir / "score_distribution_heatmap.png"),
    )

    # Print total score statistics
    print("\n" + "=" * 70)
    print("TOTAL SCORE DISTRIBUTION")
    print("=" * 70)
    for model_name in models.values():
        scores = total_scores[model_name]
        if scores:
            avg = sum(scores) / len(scores)
            print(f"{model_name}:")
            print(f"  Posts: {len(scores)}")
            print(f"  Total Score Avg: {avg:.2f}")
            print(f"  Min: {min(scores)}, Max: {max(scores)}")
            print(f"  Scores: {', '.join(map(str, sorted(scores)[:10]))}...")
        else:
            print(f"{model_name}: No data")

    # Generate total score histogram
    print("\nGenerating total score histogram...")
    plot_total_score_histogram(
        score_data=total_scores,
        output_path=str(plot_dir / "total_score_distribution.png"),
    )

    print("\n" + "=" * 70)
    print(f"DONE! Check the '{plot_dir}/' directory for outputs.")
    print("=" * 70)

    con.close()


if __name__ == "__main__":
    main()
