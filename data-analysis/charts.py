import matplotlib
import matplotlib.pyplot as plt
import seaborn as sns
import numpy as np
from matplotlib.patches import Patch
from typing import Optional

SCHWARTZ_CLUSTERS = {
    "Self-Transcendence": {
        "values": ["Tolerance", "Nature", "Equality", "Caring", "Responsibility"],
        "color": "#518B62",
    },
    "Conservation": {
        "values": [
            "Humility",
            "Respect",
            "Lawfulness",
            "Tradition",
            "Societal security",
            "Personal security",
        ],
        "color": "#57B1E3",
    },
    "Self-Enhancement": {
        "values": ["Reputation", "Achievement", "Power", "Wealth"],
        "color": "#EF9E30",
    },
    "Openness to Change": {
        "values": [
            "Pleasure",
            "Stimulation",
            "Independent actions",
            "Independent thoughts",
        ],
        "color": "#C676A2",
    },
}

MODEL_COLORS = {
    "GPT-4o-mini": "#189D7C",
    "Ministral-14b": "#F87F06",
    "Qwen3-14B": "#5E39D4",
}


def plot_bar_chart_faceted(
    data: dict[str, list[int]], output_path: str, title: str = None
):
    """
    Crea un faceted plot con 19 mini bar chart (5x4 grid) per un singolo modello.

    Args:
        data: {value: [count_0, ..., count_6]}
              Es: {"Humility": [33, 9, 44, ...], "Power": [...]}
        output_path: Path dove salvare il PNG
        title: Titolo del grafico (default: "Value Distribution")
    """
    if not data:
        print("No data to plot")
        return

    values = list(data.keys())
    n_values = len(values)

    # Grid: 5x4 = 20 slots (19 used)
    nrows, ncols = 5, 4
    fig, axes = plt.subplots(nrows, ncols, figsize=(16, 18))
    axes = axes.flatten()

    x = np.arange(7)  # 0-6

    # Color by model
    model_color = "#189D7C"  # Default green

    for idx, value in enumerate(values):
        ax = axes[idx]
        counts = data.get(value, [0] * 7)

        ax.bar(x, counts, width=0.8, color=model_color, edgecolor="white")

        ax.set_title(value, fontsize=9, fontweight="bold")
        ax.set_xticks(x)
        ax.set_xticklabels([str(v) for v in range(7)], fontsize=7)
        ax.set_ylabel("Count", fontsize=7)
        ax.grid(axis="y", alpha=0.3)

    # Hide unused subplots
    for idx in range(n_values, nrows * ncols):
        axes[idx].set_visible(False)

    fig.suptitle(
        title or "Value Distribution (0-6)", fontsize=14, fontweight="bold", y=0.98
    )
    plt.tight_layout(rect=(0, 0, 1, 0.96))
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Faceted chart saved to: {output_path}")


def plot_values_comparison(
    avg_values: dict[str, dict[str, float]],
    output_path: str,
    title: str = None,
    figsize: tuple[int, int] = None,
) -> None:
    """
    Genera un bar chart raggruppato per confrontare valori Schwartz medi.

    Args:
        avg_values: Dizionario {modello: {valore: media}}
                    Es: {"GPT": {"Reputation": 2.0, ...}, "Gemini": {...}}
        output_path: Path dove salvare il PNG
        title: Titolo del grafico (se None, genera automaticamente)
        figsize: Dimensioni figura (se None, calcola in base a N modelli)
    """
    models = list(avg_values.keys())
    n_models = len(models)

    # Auto-genera titolo
    if title is None:
        if len(models) <= 3:
            title = " vs ".join(models)
        else:
            title = (
                f"{models[0]} vs {models[1]} vs {models[2]} (+{len(models) - 3} more)"
            )

    # Auto-calcola dimensioni figura
    if figsize is None:
        height = 10 + (n_models * 0.8)
        figsize = (14, height)

    sns.set_theme(style="whitegrid")

    fig, ax = plt.subplots(figsize=figsize)

    bar_height = 0.5
    group_spacing = 0.2
    cluster_gap = 0

    current_y = 0
    y_ticks = []
    y_labels = []

    clusters_to_process = list(SCHWARTZ_CLUSTERS.items())[::-1]

    for cluster_name, cluster_data in clusters_to_process:
        available_vals = [
            v for v in cluster_data["values"] if v in avg_values[models[0]]
        ]
        if not available_vals:
            continue

        for val_name in available_vals[::-1]:
            for i, model in enumerate(models):
                val_score = avg_values[model].get(val_name, 0)

                pos = current_y + (i * bar_height)

                if n_models <= 2:
                    bar_color = cluster_data["color"]
                else:
                    bar_color = MODEL_COLORS.get(model, cluster_data["color"])

                ax.barh(
                    pos,
                    val_score,
                    height=bar_height,
                    color=bar_color,
                    edgecolor="white",
                    alpha=0.9,
                    label=model if current_y == 0 else "",
                )

                ax.text(
                    val_score + 0.05,
                    pos,
                    f"{val_score:.1f}",
                    va="center",
                    ha="left",
                    fontsize=9,
                    fontweight="bold",
                )

            y_ticks.append(current_y + (bar_height * (n_models - 1) / 2))
            y_labels.append(val_name)

            current_y += (n_models * bar_height) + group_spacing

        current_y += cluster_gap

    ax.set_yticks(y_ticks)
    ax.set_yticklabels(y_labels, fontweight="bold", fontsize=10, color="black")
    ax.set_xlabel("Average Value (0-6)", fontsize=11, fontweight="bold")

    if n_models == 1:
        y_line = 6.0
    else:
        y_line = 13.25

    ax.text(
        5,
        y_line + 0.4,
        "Social Focus",
        va="center",
        ha="left",
        fontweight="bold",
        fontsize=16,
        color="black",
    )
    ax.axhline(y=y_line, linestyle="--", color="black", linewidth=1.5, alpha=1.0)
    ax.text(
        5,
        y_line - 0.4,
        "Personal Focus",
        va="center",
        ha="left",
        fontweight="bold",
        fontsize=16,
        color="black",
    )
    ax.set_title(title, fontsize=16, fontweight="bold", pad=25)
    ax.set_xlim(0, 6.5)

    # Legenda modelli in alto a destra
    handles, labels = ax.get_legend_handles_labels()
    by_label = dict(zip(labels, handles))
    ax.legend(
        by_label.values(),
        by_label.keys(),
        loc="upper right",
        frameon=True,
        shadow=True,
    )

    # Legenda cluster in alto a sinistra
    cluster_legend_elements = [
        Patch(facecolor="#228B22", alpha=0.3, label="Self-Transcendence"),
        Patch(facecolor="#4169E1", alpha=0.3, label="Conservation"),
        Patch(facecolor="#FFA500", alpha=0.3, label="Self-Enhancement"),
        Patch(facecolor="#FF69B4", alpha=0.3, label="Openness to Change"),
    ]
    fig.legend(
        handles=cluster_legend_elements,
        loc="upper left",
        frameon=True,
        shadow=True,
        fontsize=9,
    )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Chart saved to: {output_path}")


def plot_costs_comparison(
    costs: dict[str, dict[str, float]],
    output_path: str,
    title: str = "Average Cost per Model (USD)",
) -> None:
    """
    Genera un bar chart per confrontare i costi medi per modello.

    Args:
        costs: Dizionario {modello: {"avg_cost": x, "total_cost": y, "num_posts": z}}
        output_path: Path dove salvare il PNG
        title: Titolo del grafico
    """
    models = list(costs.keys())
    avg_costs = [costs[m]["avg_cost"] for m in models]

    colors = [
        MODEL_COLORS.get(m, MODEL_COLORS.get(m.split("-")[0].capitalize(), "#666666"))
        for m in models
    ]

    sns.set_theme(style="whitegrid")
    fig, ax = plt.subplots(figsize=(10, 6))

    bars = ax.bar(models, avg_costs, color=colors, edgecolor="white", alpha=0.9)

    for bar, cost in zip(bars, avg_costs):
        ax.text(
            bar.get_x() + bar.get_width() / 2,
            bar.get_height() + 0.0001,
            f"${cost:.6f}",
            ha="center",
            va="bottom",
            fontsize=10,
            fontweight="bold",
        )

    ax.set_ylabel("Average Cost (USD)", fontsize=12, fontweight="bold")
    ax.set_title(title, fontsize=14, fontweight="bold", pad=20)
    ax.set_ylim(0, max(avg_costs) * 1.2)

    # Add total cost info
    for i, m in enumerate(models):
        total = costs[m].get("total_cost", 0)
        num = costs[m].get("num_posts", 0)
        ax.text(
            i,
            -max(avg_costs) * 0.1,
            f"Total: ${total:.4f}\n({num} posts)",
            ha="center",
            va="top",
            fontsize=8,
            color="gray",
        )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Cost chart saved to: {output_path}")


def plot_response_time_comparison(
    times: dict[str, dict[str, float]],
    output_path: str,
    title: str = "Average Response Time per Model (ms)",
) -> None:
    """
    Genera un bar chart per confrontare i tempi di risposta medi per modello.

    Args:
        times: Dizionario {modello: {"avg_time_ms": x, "total_time_ms": y, "num_posts": z}}
        output_path: Path dove salvare il PNG
        title: Titolo del grafico
    """
    models = list(times.keys())
    avg_times = [times[m]["avg_time_ms"] for m in models]

    colors = [
        MODEL_COLORS.get(m, MODEL_COLORS.get(m.split("-")[0].capitalize(), "#666666"))
        for m in models
    ]

    sns.set_theme(style="whitegrid")
    fig, ax = plt.subplots(figsize=(10, 6))

    bars = ax.bar(models, avg_times, color=colors, edgecolor="white", alpha=0.9)

    for bar, t in zip(bars, avg_times):
        ax.text(
            bar.get_x() + bar.get_width() / 2,
            bar.get_height() + max(avg_times) * 0.02,
            f"{t:.0f}ms",
            ha="center",
            va="bottom",
            fontsize=10,
            fontweight="bold",
        )

    ax.set_ylabel("Average Response Time (ms)", fontsize=12, fontweight="bold")
    ax.set_title(title, fontsize=14, fontweight="bold", pad=20)
    ax.set_ylim(0, max(avg_times) * 1.2)

    # Add total time info
    for i, m in enumerate(models):
        total = times[m].get("total_time_ms", 0)
        num = times[m].get("num_posts", 0)
        ax.text(
            i,
            -max(avg_times) * 0.1,
            f"Total: {total / 1000:.1f}s\n({num} posts)",
            ha="center",
            va="top",
            fontsize=8,
            color="gray",
        )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Response time chart saved to: {output_path}")


def plot_score_distribution_heatmap(
    distributions: dict[str, dict[str, dict[str, float]]],
    output_path: str,
    title: str = "Score Distribution: % of Posts with Zero Scores",
) -> None:
    """
    Generate a heatmap showing % of zero scores for each value across models.

    Args:
        distributions: {
            "Model1": {
                "Caring": {"pct_zeros": 61.1, "pct_mid": 22.2, "pct_high": 16.7},
                ...
            },
            "Model2": {...}
        }
        output_path: Where tosave the PNG
        title: Title for the plot
    """
    models = list(distributions.keys())

    # Get all values from first model (order matters)
    first_model = models[0]
    values = list(distributions[first_model].keys())

    # Create matrix: rows = values, columns = models
    # Cell value = % zeros
    data = np.zeros((len(values), len(models)))

    for j, model in enumerate(models):
        for i, value in enumerate(values):
            data[i, j] = distributions[model][value]["pct_zeros"]

    # Create figure
    fig, ax = plt.subplots(figsize=(10, 12))

    # Custom colormap: green (low zeros = good) -> yellow -> red (high zeros = bad)
    cmap = sns.diverging_palette(145, 10, as_cmap=True)  # Green to red

    # Create heatmap
    im = ax.imshow(data, cmap=cmap, aspect="auto", vmin=0, vmax=100)

    # Colorbar
    cbar = ax.figure.colorbar(im, ax=ax, fraction=0.046, pad=0.04)
    cbar.ax.set_ylabel(
        "% Posts with Zero Score",
        rotation=-90,
        va="bottom",
        fontsize=11,
        fontweight="bold",
    )

    # Set ticks
    ax.set_xticks(np.arange(len(models)))
    ax.set_yticks(np.arange(len(values)))

    # Labels
    ax.set_xticklabels(models, fontsize=11, fontweight="bold")
    ax.set_yticklabels(values, fontsize=10, fontweight="bold")

    # Rotate xlabels
    plt.setp(ax.get_xticklabels(), rotation=45, ha="right", rotation_mode="anchor")

    # Add text annotations
    for i in range(len(values)):
        for j in range(len(models)):
            pct = data[i, j]
            text = ax.text(
                j,
                i,
                f"{pct:.0f}%",
                ha="center",
                va="center",
                fontsize=9,
                fontweight="bold",
                color="white" if pct > 50 else "black",
            )

    ax.set_title(title, fontsize=14, fontweight="bold", pad=20)

    # Add interpretation note
    ax.text(
        0.5,
        -0.12,
        "More zeros = Conservative model (fewer values detected)\n"
        "Fewer zeros = Nuanced model (uses full 0-6 range)",
        transform=ax.transAxes,
        ha="center",
        fontsize=9,
        style="italic",
        color="gray",
    )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Score distribution heatmap saved to: {output_path}")


def plot_total_score_histogram(
    score_data: dict[str, list[int]],
    output_path: str,
    title: str = "Total Score Distribution by Model",
) -> None:
    """
    Generate histogram showing distribution of total scores across models.

    Args:
        score_data: {"Model1": [score1, score2, ...], "Model2": [...]}
        output_path: Where to save the PNG
        title: Title for the plot
    """
    models = list(score_data.keys())

    fig, ax = plt.subplots(figsize=(12, 7))

    # Colors for each model
    colors = ["#27AE60", "#FF6B35", "#3498DB", "#9B59B6", "#E74C3C", "#F39C12"]

    # Plot histogram for each model
    for i, model in enumerate(models):
        scores = score_data[model]
        if len(scores) == 0:
            continue

        # Calculate statistics
        avg = sum(scores) / len(scores) if scores else 0
        min_score = min(scores) if scores else 0
        max_score = max(scores) if scores else 0

        # Plot histogram with transparency
        ax.hist(
            scores,
            bins=range(0, 115, 5),  # Bins of 5 points
            alpha=0.5,
            label=f"{model} (avg: {avg:.1f}, n={len(scores)})",
            color=colors[i % len(colors)],
            edgecolor="black",
            linewidth=0.5,
        )

    ax.set_xlabel("Total Score (sum of 19 values)", fontsize=12, fontweight="bold")
    ax.set_ylabel("Number of Posts", fontsize=12, fontweight="bold")
    ax.set_title(title, fontsize=14, fontweight="bold", pad=20)
    ax.set_xlim(0, 114)

    # Add legend
    ax.legend(loc="upper right", frameon=True, fontsize=10)

    # Add grid
    ax.grid(axis="y", alpha=0.3)

    # Add interpretation note
    ax.text(
        0.5,
        -0.1,
        "Lower scores = Posts with fewer values expressed\n"
        "Higher scores = Posts with more values expressed (0-114 range)",
        transform=ax.transAxes,
        ha="center",
        fontsize=9,
        style="italic",
        color="gray",
    )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Total score histogram saved to: {output_path}")
