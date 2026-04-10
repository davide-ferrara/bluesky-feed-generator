"""Create PowerPoint-ready visualization for outlier analysis."""

import sqlite3
import matplotlib.pyplot as plt
import numpy as np
from pathlib import Path


def load_score_data(cur: sqlite3.Cursor) -> dict:
    """Load total scores for all models."""

    db_columns = [
        "reputation",
        "power",
        "wealth",
        "achievement",
        "pleasure",
        "independent_thoughts",
        "independent_actions",
        "stimulation",
        "personal_security",
        "societal_security",
        "tradition",
        "lawfulness",
        "respect",
        "humility",
        "responsibility",
        "caring",
        "equality",
        "nature",
        "tolerance",
    ]

    sum_expr = " + ".join(db_columns)

    models = {
        "openai/gpt-4o-mini": "GPT-4o-mini",
        "mistralai/ministral-8b-2512": "Ministral-8b",
    }

    score_data = {}

    for model_name, display_name in models.items():
        query = f"""
            SELECT {sum_expr} as total_score
            FROM analyses
            WHERE model = ?
            ORDER BY total_score
        """
        cur.execute(query, (model_name,))
        scores = [row[0] for row in cur.fetchall() if row[0] is not None]
        if scores:  # Only include models with data
            score_data[display_name] = scores

    return score_data


def create_boxplot_scatter(score_data: dict, output_path: str):
    """Create publication-ready box plot with scatter overlay."""

    if not score_data:
        print("Warning: No data available for visualization")
        return

    # Setup for high-quality PPT export
    plt.style.use("seaborn-v0_8-whitegrid")
    fig, ax = plt.subplots(figsize=(12, 7))

    models = list(score_data.keys())

    # Create box plot
    bp = ax.boxplot(
        [score_data[model] for model in models],
        labels=[f"{model}\n(V4)" for model in models],
        patch_artist=True,
        widths=0.5,
        notch=False,
        showfliers=False,  # We'll add our own scatter
    )

    # Colors
    colors = ["#27AE60", "#FF6B35"]  # Green for GPT, Orange for Ministral

    for patch, color in zip(bp["boxes"], colors):
        patch.set_facecolor(color)
        patch.set_alpha(0.4)
        patch.set_edgecolor("black")
        patch.set_linewidth(2)

    # Style whiskers, caps, medians
    for whisker in bp["whiskers"]:
        whisker.set(color="#333333", linewidth=1.5)
    for cap in bp["caps"]:
        cap.set(color="#333333", linewidth=1.5)
    for median in bp["medians"]:
        median.set(color="black", linewidth=2.5)

    # Add scatter points with jitter
    for i, (model, scores) in enumerate(score_data.items()):
        # Calculate statistics
        mean = np.mean(scores)
        std = np.std(scores)

        # Jitter for visibility
        jitter = np.random.normal(0, 0.04, len(scores))
        x = np.array([i + 1] * len(scores)) + jitter

        # Plot all points
        ax.scatter(
            x,
            scores,
            c=colors[i],
            alpha=0.6,
            s=80,
            edgecolors="black",
            linewidths=0.5,
            zorder=3,
        )

        # Identify outliers (2 std from mean)
        outlier_threshold_high = mean + 2 * std
        outlier_threshold_low = max(0, mean - 2 * std)

        # Annotate high outliers
        outliers = [
            (idx, val)
            for idx, val in enumerate(scores)
            if val > outlier_threshold_high or val < outlier_threshold_low
        ]

        if outliers:
            # Sort by score (highest first) for labelling
            outliers_sorted = sorted(outliers, key=lambda x: x[1], reverse=True)[:3]

            for idx, val in outliers_sorted:
                # Offset labels to avoid overlap
                offset_x = 0.15 if i == 0 else -0.15
                offset_y = 8 if val > mean else -8

                ax.annotate(
                    f"{int(val)}",
                    (i + 1, val),
                    xytext=(offset_x, offset_y),
                    textcoords="offset points",
                    ha="center",
                    fontsize=11,
                    fontweight="bold",
                    color=colors[i],
                    bbox=dict(
                        boxstyle="round,pad=0.3",
                        facecolor="white",
                        edgecolor=colors[i],
                        alpha=0.9,
                    ),
                    zorder=5,
                )

    # Add statistics box
    stats_text = []
    for i, model in enumerate(models):
        scores = score_data[model]
        mean = np.mean(scores)
        std = np.std(scores)
        median = np.median(scores)
        min_val = np.min(scores)
        max_val = np.max(scores)

        stats_text.append(
            f"{model}:\n"
            f"  Mean: {mean:.1f}\n"
            f"  Median: {median:.1f}\n"
            f"  Std: {std:.1f}\n"
            f"  Range: {min_val}-{max_val}"
        )

    # Calculate improvement (only if we have 2+ models)
    if len(models) >= 2:
        gpt_mean = np.mean(score_data[models[0]])
        min_mean = np.mean(score_data[models[1]])
        improvement = ((min_mean / gpt_mean) - 1) * 100
        stats_box = "\n\n".join(stats_text) + f"\n\nImprovement: +{improvement:.0f}%"
    else:
        stats_box = "\n\n".join(stats_text)

    plt.text(
        0.02,
        0.98,
        stats_box,
        transform=ax.transAxes,
        fontsize=10,
        verticalalignment="top",
        fontfamily="monospace",
        bbox=dict(boxstyle="round", facecolor="lightblue", edgecolor="navy", alpha=0.3),
    )

    # Highlight improvement callout (only if we have 2+ models)
    if len(models) >= 2:
        plt.text(
            0.98,
            0.02,
            f"Ministral-8b: +{improvement:.0f}% higher average",
            transform=ax.transAxes,
            fontsize=11,
            fontweight="bold",
            ha="right",
            va="bottom",
            bbox=dict(
                boxstyle="round",
                facecolor="lightgreen",
                edgecolor="darkgreen",
                alpha=0.5,
            ),
        )

    # Styling
    ax.set_ylabel(
        "Total Score\n(sum of 19 values, max: 114)", fontsize=12, fontweight="bold"
    )
    ax.set_xlabel("Model", fontsize=11, fontweight="bold")
    ax.set_title(
        "PROMPT V4: Outlier Analysis & Score Distribution\n"
        "Higher Scores = More Values Detected per Post",
        fontsize=14,
        fontweight="bold",
        pad=20,
    )

    ax.set_ylim(0, 45)
    ax.set_xlim(0.5, 2.5)
    ax.grid(axis="y", alpha=0.3, linestyle="--")
    ax.set_axisbelow(True)

    # Add reference line at mean
    for i, model in enumerate(models):
        mean = np.mean(score_data[model])
        ax.axhline(
            y=mean,
            xmin=i / 2 + 0.15,
            xmax=(i + 1) / 2 - 0.15,
            color=colors[i],
            linestyle="--",
            linewidth=2,
            alpha=0.7,
        )

    plt.tight_layout()
    plt.savefig(
        output_path, dpi=300, bbox_inches="tight", facecolor="white", edgecolor="none"
    )
    plt.close()

    print(f"\n✓ Box plot saved: {output_path}")
    print(f"  Resolution: 300 DPI (PowerPoint ready)")
    print(f"  Background: White (transparent in PPT)")

    return output_path


def create_summary_text(score_data: dict) -> str:
    """Create text summary for PPT notes."""

    summary = []
    summary.append("=" * 70)
    summary.append("OUTLIER ANALYSIS SUMMARY")
    summary.append("=" * 70)
    summary.append("\nKey Findings:")
    summary.append("-" * 70)

    for model, scores in score_data.items():
        mean = np.mean(scores)
        std = np.std(scores)
        median = np.median(scores)

        summary.append(f"\n{model}:")
        summary.append(f"  Average: {mean:.1f}")
        summary.append(f"  Median: {median:.1f}")
        summary.append(f"  Std Dev: {std:.1f}")
        summary.append(f"  Range: {min(scores)} - {max(scores)}")

        # Outliers
        outliers_high = [s for s in scores if s > mean + 2 * std]
        outliers_low = [s for s in scores if s < max(0, mean - 2 * std)]

        if outliers_high:
            summary.append(f"  High outliers: {sorted(outliers_high, reverse=True)}")
        if outliers_low:
            summary.append(f"  Low outliers: {sorted(outliers_low)}")

    # Comparison (only if we have 2+ models)
    models = list(score_data.keys())
    if len(models) >= 2:
        improvement = (
            (np.mean(score_data[models[1]]) / np.mean(score_data[models[0]])) - 1
        ) * 100
        summary.append("\n" + "-" * 70)
        summary.append(f"\nComparison:")
        summary.append(f"  {models[1]} scores {improvement:.0f}% higher on average")
        summary.append(f"  Wider range indicates better variance detection")
        summary.append(f"  Outliers represent posts with rich value content")
    else:
        summary.append("\n" + "-" * 70)
        summary.append(f"\nSingle model comparison not available")

    return "\n".join(summary)


def main():
    # Connect to database
    con = sqlite3.connect("../data.db")
    cur = con.cursor()

    # Create output directory
    output_dir = Path("outliers")
    output_dir.mkdir(exist_ok=True)

    # Load data
    print("Loading score data...")
    score_data = load_score_data(cur)

    # Create visualization
    print("\nCreating box plot visualization...")
    output_path = output_dir / "outliers_boxplot_scatter.png"
    create_boxplot_scatter(score_data, str(output_path))

    # Create summary
    summary = create_summary_text(score_data)
    print("\n" + summary)

    # Save summary
    summary_path = output_dir / "outliers_summary.txt"
    with open(summary_path, "w") as f:
        f.write(summary)
    print(f"\n✓ Summary saved: {summary_path}")

    con.close()

    print("\n" + "=" * 70)
    print("READY FOR POWERPOINT!")
    print("=" * 70)
    print(f"\nFiles created:")
    print(f"  1. {output_path} (main visualization)")
    print(f"  2. {summary_path} (notes for speaker)")
    print("\nUsage in PowerPoint:")
    print("  - Insert PNG image (high quality, 300 DPI)")
    print("  - Add title: 'Model Comparison: Outlier Analysis'")
    print("  - Use summary.txt for speaker notes")


if __name__ == "__main__":
    main()
