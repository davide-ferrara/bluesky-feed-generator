import json
import os
import glob
from collections import Counter
from charts import (
    plot_values_comparison,
    plot_costs_comparison,
    plot_response_time_comparison,
)
from spider import plot_spider_chart


def find_latest_json(pattern: str) -> str:
    """Find the most recent JSON file matching the pattern."""
    files = glob.glob(pattern)
    if not files:
        return None
    # Sort by modification time, most recent first
    files.sort(key=os.path.getmtime, reverse=True)
    return files[0]


def load_json(file_path: str):
    if not file_path or not file_path.endswith("json"):
        return None

    with open(file_path) as f:
        data = json.load(f)

    return data


def calculate_values_avg(posts):
    """Calculate average values for each Schwartz dimension."""
    if not posts:
        return {}

    # Get all value names from first post
    value_names = list(posts[0]["ValueAnalysis"]["Rating"].keys())

    avg_values = {}
    for name in value_names:
        total = sum(post["ValueAnalysis"]["Rating"][name] for post in posts)
        avg_values[name] = total / len(posts)

    return avg_values


def calculate_values_mode(posts):
    """Calculate mode (most frequent value) for each Schwartz dimension."""
    if not posts:
        return {}

    # Get all value names from first post
    value_names = list(posts[0]["ValueAnalysis"]["Rating"].keys())

    mode_values = {}
    for name in value_names:
        # Collect all values for this dimension
        all_values = [post["ValueAnalysis"]["Rating"][name] for post in posts]
        # Find most common value (mode)
        counter = Counter(all_values)
        mode_values[name] = counter.most_common(1)[0][0]

    return mode_values


def calculate_cost_stats(posts):
    """Calculate average and total cost for model."""
    if not posts:
        return {}

    total_cost = 0
    for post in posts:
        total_cost += post["ValueAnalysis"]["Stats"]["cost_usd"]

    avg_cost = total_cost / len(posts)

    return {
        "avg_cost": avg_cost,
        "total_cost": total_cost,
        "num_posts": len(posts),
    }


def calculate_time_stats(posts):
    """Calculate average and total response time for model."""
    if not posts:
        return {}

    total_time = 0
    for post in posts:
        total_time += post["ValueAnalysis"]["Stats"]["response_time_ms"]

    avg_time = total_time / len(posts)

    return {
        "avg_time_ms": avg_time,
        "total_time_ms": total_time,
        "num_posts": len(posts),
    }


def main():
    # Trova automaticamente i file JSON più recenti per ogni modello
    gpt_file = find_latest_json("../post_gpt-4.1-mini*.json")
    mistral_file = find_latest_json("../post_ministral-14b*.json")
    deepseek_file = find_latest_json("../post_DeepSeek-V3*.json")
    qwen_file = find_latest_json("../post_Qwen3-VL*.json")

    print(f"GPT file: {gpt_file}")
    print(f"Mistral file: {mistral_file}")
    print(f"DeepSeek file: {deepseek_file}")
    print(f"Qwen file: {qwen_file}")

    # Carica i dati
    gpt_posts = load_json(gpt_file)
    mistral_posts = load_json(mistral_file)
    deepseek_posts = load_json(deepseek_file)
    qwen_posts = load_json(qwen_file)

    if not gpt_posts or not mistral_posts or not deepseek_posts or not qwen_posts:
        print("Error loading data")
        if not gpt_posts:
            print("  - GPT posts not found")
        if not mistral_posts:
            print("  - Mistral posts not found")
        if not deepseek_posts:
            print("  - DeepSeek posts not found")
        if not qwen_posts:
            print("  - Qwen posts not found")
        return

    print("\nData loaded successfully")
    print(f"GPT posts: {len(gpt_posts)}")
    print(f"Mistral posts: {len(mistral_posts)}")
    print(f"DeepSeek posts: {len(deepseek_posts)}")
    print(f"Qwen posts: {len(qwen_posts)}")

    # Calcola le medie per ogni modello
    gpt_avg_values = calculate_values_avg(gpt_posts)
    mistral_avg_values = calculate_values_avg(mistral_posts)
    deepseek_avg_values = calculate_values_avg(deepseek_posts)
    qwen_avg_values = calculate_values_avg(qwen_posts)

    # Calcola le mode per ogni modello
    gpt_mode_values = calculate_values_mode(gpt_posts)
    mistral_mode_values = calculate_values_mode(mistral_posts)
    deepseek_mode_values = calculate_values_mode(deepseek_posts)
    qwen_mode_values = calculate_values_mode(qwen_posts)

    print("GPT averages:", gpt_avg_values)
    print("\nMistral averages:", mistral_avg_values)
    print("\nDeepSeek averages:", deepseek_avg_values)
    print("\nQwen averages:", qwen_avg_values)

    print("\nGPT mode:", gpt_mode_values)
    print("\nMistral mode:", mistral_mode_values)
    print("\nDeepSeek mode:", deepseek_mode_values)
    print("\nQwen mode:", qwen_mode_values)

    # Calcola statistiche costi e tempi
    gpt_cost = calculate_cost_stats(gpt_posts)
    mistral_cost = calculate_cost_stats(mistral_posts)
    deepseek_cost = calculate_cost_stats(deepseek_posts)
    qwen_cost = calculate_cost_stats(qwen_posts)

    gpt_time = calculate_time_stats(gpt_posts)
    mistral_time = calculate_time_stats(mistral_posts)
    deepseek_time = calculate_time_stats(deepseek_posts)
    qwen_time = calculate_time_stats(qwen_posts)

    print("\n" + "=" * 50)
    print("COST STATISTICS")
    print("=" * 50)
    print(f"GPT: avg=${gpt_cost['avg_cost']:.6f}, total=${gpt_cost['total_cost']:.4f}")
    print(
        f"Mistral: avg=${mistral_cost['avg_cost']:.6f}, total=${mistral_cost['total_cost']:.4f}"
    )
    print(
        f"DeepSeek: avg=${deepseek_cost['avg_cost']:.6f}, total=${deepseek_cost['total_cost']:.4f}"
    )
    print(
        f"Qwen: avg=${qwen_cost['avg_cost']:.6f}, total=${qwen_cost['total_cost']:.4f}"
    )

    print("\n" + "=" * 50)
    print("RESPONSE TIME STATISTICS")
    print("=" * 50)
    print(
        f"GPT: avg={gpt_time['avg_time_ms']:.0f}ms, total={gpt_time['total_time_ms'] / 1000:.1f}s"
    )
    print(
        f"Mistral: avg={mistral_time['avg_time_ms']:.0f}ms, total={mistral_time['total_time_ms'] / 1000:.1f}s"
    )
    print(
        f"DeepSeek: avg={deepseek_time['avg_time_ms']:.0f}ms, total={deepseek_time['total_time_ms'] / 1000:.1f}s"
    )
    print(
        f"Qwen: avg={qwen_time['avg_time_ms']:.0f}ms, total={qwen_time['total_time_ms'] / 1000:.1f}s"
    )

    # Spider plot individuali
    # plot_spider_chart(
    #     avg_values={"GPT-4.1-mini": gpt_avg_values},
    #     output_path="spider_gpt.png",
    #     title="GPT-4.1-mini - Schwartz Values",
    # )

    # plot_spider_chart(
    #     avg_values={"Mistral-14b": mistral_avg_values},
    #     output_path="spider_mistral.png",
    #     title="Mistral-14b - Schwartz Values",
    # )

    # plot_spider_chart(
    #     avg_values={"DeepSeek": deepseek_avg_values},
    #     output_path="spider_deepseek.png",
    #     title="DeepSeek - Schwartz Values",
    # )

    # plot_spider_chart(
    #     avg_values={"Qwen3": qwen_avg_values},
    #     output_path="spider_qwen.png",
    #     title="Qwen3 - Schwartz Values",
    # )

    # Bar chart individuali
    plot_values_comparison(
        avg_values={"GPT-4.1-mini": gpt_avg_values},
        output_path="bar_gpt.png",
    )

    plot_values_comparison(
        avg_values={"Mistral-14b": mistral_avg_values},
        output_path="bar_mistral.png",
    )

    plot_values_comparison(
        avg_values={"DeepSeek": deepseek_avg_values},
        output_path="bar_deepseek.png",
    )

    plot_values_comparison(
        avg_values={"Qwen3": qwen_avg_values},
        output_path="bar_qwen.png",
    )

    # Spider plot combinato
    # plot_spider_chart(
    #     avg_values={
    #         "GPT-4.1-mini": gpt_avg_values,
    #         "Mistral-14b": mistral_avg_values,
    #         "DeepSeek": deepseek_avg_values,
    #         "Qwen3": qwen_avg_values,
    #     },
    #     output_path="spider_4_models.png",
    # )

    # Bar chart combinato (average)
    plot_values_comparison(
        avg_values={
            "GPT-4.1-mini": gpt_avg_values,
            "Mistral-14b": mistral_avg_values,
            "DeepSeek": deepseek_avg_values,
            "Qwen3": qwen_avg_values,
        },
        output_path="comparison_4_models.png",
    )

    # =====================================
    # MODE PLOTS
    # =====================================

    print("\n" + "=" * 50)
    print("Generating MODE plots...")
    print("=" * 50)

    # Spider plot individuali (mode)
    # plot_spider_chart(
    #     avg_values={"GPT-4.1-mini": gpt_mode_values},
    #     output_path="spider_gpt_mode.png",
    #     title="GPT-4.1-mini - Schwartz Values (Mode)",
    # )

    # plot_spider_chart(
    #     avg_values={"Mistral-14b": mistral_mode_values},
    #     output_path="spider_mistral_mode.png",
    #     title="Mistral-14b - Schwartz Values (Mode)",
    # )

    # plot_spider_chart(
    #     avg_values={"DeepSeek": deepseek_mode_values},
    #     output_path="spider_deepseek_mode.png",
    #     title="DeepSeek - Schwartz Values (Mode)",
    # )

    # plot_spider_chart(
    #     avg_values={"Qwen3": qwen_mode_values},
    #     output_path="spider_qwen_mode.png",
    #     title="Qwen3 - Schwartz Values (Mode)",
    # )

    # Bar chart individuali (mode)
    plot_values_comparison(
        avg_values={"GPT-4.1-mini": gpt_mode_values},
        output_path="bar_gpt_mode.png",
    )

    plot_values_comparison(
        avg_values={"Mistral-14b": mistral_mode_values},
        output_path="bar_mistral_mode.png",
    )

    plot_values_comparison(
        avg_values={"DeepSeek": deepseek_mode_values},
        output_path="bar_deepseek_mode.png",
    )

    plot_values_comparison(
        avg_values={"Qwen3": qwen_mode_values},
        output_path="bar_qwen_mode.png",
    )

    # Spider plot combinato (mode)
    # plot_spider_chart(
    #     avg_values={
    #         "GPT-4.1-mini": gpt_mode_values,
    #         "Mistral-14b": mistral_mode_values,
    #         "DeepSeek": deepseek_mode_values,
    #         "Qwen3": qwen_mode_values,
    #     },
    #     output_path="spider_4_models_mode.png",
    # )

    # Bar chart combinato (mode)
    plot_values_comparison(
        avg_values={
            "GPT-4.1-mini": gpt_mode_values,
            "Mistral-14b": mistral_mode_values,
            "DeepSeek": deepseek_mode_values,
            "Qwen3": qwen_mode_values,
        },
        output_path="comparison_4_models_mode.png",
    )

    # =====================================
    # COST AND TIME PLOTS
    # =====================================

    print("\n" + "=" * 50)
    print("Generating cost and time plots...")
    print("=" * 50)

    # Cost comparison plot
    plot_costs_comparison(
        costs={
            "GPT-4.1-mini": gpt_cost,
            "Mistral-14b": mistral_cost,
            "DeepSeek": deepseek_cost,
            "Qwen3": qwen_cost,
        },
        output_path="comparison_costs.png",
    )

    # Response time comparison plot
    plot_response_time_comparison(
        times={
            "GPT-4.1-mini": gpt_time,
            "Mistral-14b": mistral_time,
            "DeepSeek": deepseek_time,
            "Qwen3": qwen_time,
        },
        output_path="comparison_response_time.png",
    )


if __name__ == "__main__":
    main()
