"""Statistics calculation functions from SQL database."""

import sqlite3

# def count_values(cur: sqlite3.Cursor, model: str) -> dict[str, float]:
def count_values(cur: sqlite3.Cursor, model: str):
    """Count values occurence for each post from SQL database."""
    # Database column names (snake_case) - order matches Go code
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

    # Display names expected by plot function (Title Case)
    display_names = {
        "reputation": "Reputation",
        "power": "Power",
        "wealth": "Wealth",
        "achievement": "Achievement",
        "pleasure": "Pleasure",
        "independent_thoughts": "Independent thoughts",
        "independent_actions": "Independent actions",
        "stimulation": "Stimulation",
        "personal_security": "Personal security",
        "societal_security": "Societal security",
        "tradition": "Tradition",
        "lawfulness": "Lawfulness",
        "respect": "Respect",
        "humility": "Humility",
        "responsibility": "Responsibility",
        "caring": "Caring",
        "equality": "Equality",
        "nature": "Nature",
        "tolerance": "Tolerance",
    }

    data = dict()

    # Build query
    value_distribution = []
    for value in db_columns:
        res = []
        for i in range(0, 6+1):

            query = f"SELECT COUNT(*), a.{value} FROM ANALYSES AS a WHERE model=? AND a.{value} = {i}"
            cur.execute(query, (model,))
            row = cur.fetchone()
            res.append(row[0])

        data[display_names[value]] = res
    value_distribution.append(data)

    return value_distribution


def calculate_values_avg(cur: sqlite3.Cursor, model: str) -> dict[str, float]:
    """Calculate average values for each Schwartz dimension from SQL database."""
    # Database column names (snake_case) - order matches Go code
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

    # Display names expected by plot function (Title Case)
    display_names = {
        "reputation": "Reputation",
        "power": "Power",
        "wealth": "Wealth",
        "achievement": "Achievement",
        "pleasure": "Pleasure",
        "independent_thoughts": "Independent thoughts",
        "independent_actions": "Independent actions",
        "stimulation": "Stimulation",
        "personal_security": "Personal security",
        "societal_security": "Societal security",
        "tradition": "Tradition",
        "lawfulness": "Lawfulness",
        "respect": "Respect",
        "humility": "Humility",
        "responsibility": "Responsibility",
        "caring": "Caring",
        "equality": "Equality",
        "nature": "Nature",
        "tolerance": "Tolerance",
    }

    # Build query with AVG for each value column
    select_cols = ", ".join([f"AVG({v}) as {v}" for v in db_columns])
    query = f"SELECT {select_cols} FROM ANALYSES WHERE model=?"
    cur.execute(query, (model,))
    row = cur.fetchone()

    # Convert to dict with display names
    avg_values = {}
    for i, db_name in enumerate(db_columns):
        display_name = display_names[db_name]
        avg_values[display_name] = row[i] if row[i] is not None else 0

    return avg_values


def calculate_score_distribution(
    cur: sqlite3.Cursor, model: str
) -> dict[str, dict[str, float]]:
    """
    Calculate distribution of scores (0, 1-3, 4-6) for each value.

    Returns:
        {
            "Caring": {"pct_zeros": 61.1, "pct_mid": 22.2, "pct_high": 16.7},
            "Equality": {"pct_zeros": 66.7, "pct_mid": 11.1, "pct_high": 22.2},
            ...
        }
    """
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

    display_names = {
        "reputation": "Reputation",
        "power": "Power",
        "wealth": "Wealth",
        "achievement": "Achievement",
        "pleasure": "Pleasure",
        "independent_thoughts": "Independent thoughts",
        "independent_actions": "Independent actions",
        "stimulation": "Stimulation",
        "personal_security": "Personal security",
        "societal_security": "Societal security",
        "tradition": "Tradition",
        "lawfulness": "Lawfulness",
        "respect": "Respect",
        "humility": "Humility",
        "responsibility": "Responsibility",
        "caring": "Caring",
        "equality": "Equality",
        "nature": "Nature",
        "tolerance": "Tolerance",
    }

    distribution = {}

    for db_col in db_columns:
        query = f"""
            SELECT 
                COUNT(*) as total,
                SUM(CASE WHEN {db_col} = 0 THEN 1 ELSE 0 END) as zeros,
                SUM(CASE WHEN {db_col} BETWEEN 1 AND 3 THEN 1 ELSE 0 END) as mid_range,
                SUM(CASE WHEN {db_col} >= 4 THEN 1 ELSE 0 END) as high_scores
            FROM analyses
            WHERE model=?
        """
        cur.execute(query, (model,))
        row = cur.fetchone()

        total = row[0] if row[0] else 1  # Avoid division by zero
        zeros = row[1] if row[1] else 0
        mid_range = row[2] if row[2] else 0
        high_scores = row[3] if row[3] else 0

        display_name = display_names[db_col]
        distribution[display_name] = {
            "pct_zeros": round(zeros * 100.0 / total, 1),
            "pct_mid": round(mid_range * 100.0 / total, 1),
            "pct_high": round(high_scores * 100.0 / total, 1),
        }

    return distribution


def calculate_cost_stats(cur: sqlite3.Cursor, model: str) -> dict[str, float]:
    """Calculate average and total cost for model from SQL database."""
    query = """
        SELECT 
            AVG(json_extract(stats, '$.cost_usd')) as avg_cost,
            SUM(json_extract(stats, '$.cost_usd')) as total_cost,
            COUNT(*) as num_posts
        FROM ANALYSES
        WHERE model=? AND stats IS NOT NULL
    """
    cur.execute(query, (model,))
    row = cur.fetchone()

    return {
        "avg_cost": row[0] if row[0] is not None else 0,
        "total_cost": row[1] if row[1] is not None else 0,
        "num_posts": row[2] if row[2] is not None else 0,
    }


def calculate_time_stats(cur: sqlite3.Cursor, model: str) -> dict[str, float]:
    """Calculate average and total response time for model from SQL database."""
    query = """
        SELECT 
            AVG(json_extract(stats, '$.response_time_ms')) as avg_time,
            SUM(json_extract(stats, '$.response_time_ms')) as total_time,
            COUNT(*) as num_posts
        FROM ANALYSES
        WHERE model=? AND stats IS NOT NULL
    """
    cur.execute(query, (model,))
    row = cur.fetchone()

    return {
        "avg_time_ms": row[0] if row[0] is not None else 0,
        "total_time_ms": row[1] if row[1] is not None else 0,
        "num_posts": row[2] if row[2] is not None else 0,
    }


def calculate_total_score_distribution(cur: sqlite3.Cursor, model: str) -> list[int]:
    """
    Calculate total score (sum of all 19 values) for each post.
    
    Returns:
        List of total scores for each post by this model
    """
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
    
    # Build query to sum all values
    sum_expr = " + ".join(db_columns)
    query = f"""
        SELECT {sum_expr} as total_score
        FROM analyses
        WHERE model = ?
        ORDER BY total_score
    """
    cur.execute(query, (model,))
    rows = cur.fetchall()
    
    # Return list of total scores (filter out None)
    return [row[0] for row in rows if row[0] is not None]
