"""Export functions forCSV and plots."""

import csv
from pathlib import Path

from charts import (
    plot_values_comparison,
    plot_costs_comparison,
    plot_response_time_comparison,
)


def save_to_csv(stats: dict, plot_dir: Path) -> None:
    """Save statistics to CSV files for easy reading."""
    # Get all value names from first model
    first_model = list(stats.keys())[0]
    value_names = list(stats[first_model]["avg_values"].keys())

    # Save average values
    with open(plot_dir / "values_avg.csv", "w", newline="") as f:
        writer = csv.writer(f)
        # Header
        writer.writerow(["Model"] + value_names)
        # Data rows
        for model_name, data in stats.items():
            row = [model_name] + [data["avg_values"][v] for v in value_names]
            writer.writerow(row)

    # Save cost statistics
    with open(plot_dir / "costs.csv", "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["Model", "Avg Cost (USD)", "Total Cost (USD)", "Num Posts"])
        for model_name, data in stats.items():
            row = [
                model_name,
                data["cost_stats"]["avg_cost"],
                data["cost_stats"]["total_cost"],
                data["cost_stats"]["num_posts"],
            ]
            writer.writerow(row)

    # Save response time statistics
    with open(plot_dir / "response_times.csv", "w", newline="") as f:
        writer = csv.writer(f)
        writer.writerow(["Model", "Avg Time (ms)", "Total Time (ms)", "Num Posts"])
        for model_name, data in stats.items():
            row = [
                model_name,
                data["time_stats"]["avg_time_ms"],
                data["time_stats"]["total_time_ms"],
                data["time_stats"]["num_posts"],
            ]
            writer.writerow(row)

    print(f"\nCSV files saved to: {plot_dir}/")


def save_distribution_csv(distributions: dict, plot_dir: Path) -> None:
    """Save score distribution to CSV."""
    # Get all value names from first model
    first_model = list(distributions.keys())[0]
    value_names = list(distributions[first_model].keys())

    # Save distribution data
    with open(plot_dir / "score_distribution.csv", "w", newline="") as f:
        writer = csv.writer(f)
        # Header: Model, Value, % Zeros, % Mid, % High
        writer.writerow(["Model", "Value", "% Zeros", "% Mid (1-3)", "% High (4-6)"])
        # Data rows
        for model_name, model_data in distributions.items():
            for value_name in value_names:
                dist = model_data[value_name]
                row = [
                    model_name,
                    value_name,
                    dist["pct_zeros"],
                    dist["pct_mid"],
                    dist["pct_high"],
                ]
                writer.writerow(row)

    print(f"Distribution CSV saved to: {plot_dir}/score_distribution.csv")


def generate_plots(stats: dict, model_names: list[str], plot_dir: Path) -> None:
    """Generate all comparison plots."""
    # Bar charts for each model (average)
    for name in model_names:
        plot_values_comparison(
            avg_values={name: stats[name]["avg_values"]},
            output_path=str(plot_dir / f"bar_{name.replace('-', '_').lower()}.png"),
        )

    # Combined comparison (average)
    plot_values_comparison(
        avg_values={name: stats[name]["avg_values"] for name in model_names},
        output_path=str(plot_dir / "comparison_avg.png"),
    )

    # Cost and time comparison
    plot_costs_comparison(
        costs={name: stats[name]["cost_stats"] for name in model_names},
        output_path=str(plot_dir / "comparison_costs.png"),
    )

    plot_response_time_comparison(
        times={name: stats[name]["time_stats"] for name in model_names},
        output_path=str(plot_dir / "comparison_response_time.png"),
    )
