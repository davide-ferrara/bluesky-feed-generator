import matplotlib.pyplot as plt
import numpy as np
from typing import Optional

SCHWARTZ_VALUES = [
    "Independent thoughts",
    "Independent actions",
    "Stimulation",
    "Pleasure",
    "Achievement",
    "Power",
    "Wealth",
    "Reputation",
    "Personal security",
    "Societal security",
    "Tradition",
    "Lawfulness",
    "Respect",
    "Humility",
    "Caring",
    "Responsibility",
    "Equality",
    "Nature",
    "Tolerance",
]

MODEL_COLORS = {
    "GPT": "#27AE60",
    "GPT-4.1-mini": "#27AE60",
    "Mistral": "#FF6B35",
    "Mistral-14b": "#FF6B35",
    "ministral-14b": "#FF6B35",
    "DeepSeek": "#2E86AB",
    "DeepSeek-V3": "#2E86AB",
    "deepseek-v3": "#2E86AB",
    "Qwen": "#8E44AD",
    "Qwen3": "#8E44AD",
    "Qwen3-VL-30B": "#8E44AD",
}


def plot_spider_chart(
    avg_values: dict[str, dict[str, float]],
    output_path: str,
    title: str = None,
    figsize: tuple[int, int] = None,
) -> None:
    """
    Genera un radar/spider chart per confrontare i 19 valori Schwartz.

    Args:
        avg_values: Dizionario {modello: {valore: media}}
                    Es: {"GPT": {"Reputation": 2.0, ...}, "Gemini": {...}}
        output_path: Path dove salvare il PNG
        title: Titolo del grafico (se None, genera automaticamente)
        figsize: Dimensioni figura (se None, calcola in base a N modelli)
    """
    models = list(avg_values.keys())
    n_models = len(models)
    n_values = len(SCHWARTZ_VALUES)

    if title is None:
        if n_models == 1:
            title = f"{models[0]} - Schwartz Values"
        elif n_models <= 3:
            title = " vs ".join(models)
        else:
            title = f"{n_models} Models Comparison"

    if figsize is None:
        figsize = (10, 10)

    angles = np.linspace(0, 2 * np.pi, n_values, endpoint=False).tolist()
    angles += angles[:1]

    fig, ax = plt.subplots(figsize=figsize, subplot_kw=dict(polar=True))

    for i, model in enumerate(models):
        model_color = MODEL_COLORS.get(model)
        if model_color is None:
            first_part = model.split("-")[0]
            model_color = MODEL_COLORS.get(first_part.capitalize(), f"C{i}")

        values = []
        for val in SCHWARTZ_VALUES:
            values.append(avg_values[model].get(val, 0))
        values += values[:1]

        ax.plot(
            angles,
            values,
            linewidth=2,
            linestyle="solid",
            label=model,
            color=model_color,
        )
        ax.fill(angles, values, alpha=0.15, color=model_color)

    ax.set_xticks(angles[:-1])
    ax.set_xticklabels(SCHWARTZ_VALUES, fontsize=9, fontweight="bold")

    ax.set_ylim(0, 6)
    ax.set_yticks([1, 2, 3, 4, 5, 6])
    ax.set_yticklabels(["1", "2", "3", "4", "5", "6"], fontsize=8)

    ax.set_title(title, fontsize=16, fontweight="bold", pad=25)

    ax.legend(
        loc="upper right",
        bbox_to_anchor=(1.15, 1.1),
        frameon=True,
        shadow=True,
        fontsize=10,
    )

    plt.tight_layout()
    plt.savefig(output_path, dpi=150, bbox_inches="tight")
    plt.close()

    print(f"Spider chart saved to: {output_path}")
