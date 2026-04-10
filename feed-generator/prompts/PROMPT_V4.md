# Task: Analyze Values in Social Media Posts

Answer like the expert sociologist Barry Schwartz, your task is to identify how strongly each value is expressed or implied in the text.

## Definitions

1. INDEPENDENT THOUGHT: Freedom to cultivate one's own ideas and abilities
2. INDEPENDENT ACTION: Freedom to determine one's own actions
3. STIMULATION: Excitement, novelty, and change
4. PLEASURE: Enjoyment and hedonistic pursuits
5. ACHIEVEMENT: Success according to social standards
6. POWER: Influence and authority over others
7. WEALTH: Control of material and social resources
8. REPUTATION: Maintaining one's public image and avoiding humiliation
9. PERSONAL SECURITY: Safety in one's immediate environment
10. SOCIETAL SECURITY: Safety and stability in the wider society
11. TRADITION: Maintaining cultural, family, or religious traditions
12. LAWFULNESS: Compliance with rules, laws, and formal obligations
13. RESPECT: Avoiding upsetting or harming other people
14. HUMILITY: Being humble and modest
15. CARING: Devotion to the welfare of people they care about
16. RESPONSIBILITY: Being responsible and having loyalty to others
17. EQUALITY: Commitment to equality, justice, and protection for all people
18. NATURE: Preservation of the natural environment
19. TOLERANCE: Acceptance of those different from oneself

## Important Guidelines

1. **Use the FULL range 0-6 freely**
2. Consider BOTH explicit statements AND implicit values (tone, context, implications)
3. Use mid-range scores (2-4) for:
   - Subtle or moderate expressions of values
   - Values implied by context or tone
   - Values reflected indirectly in the content
4. Score 0 ONLY when a value is genuinely absent or actively contradicted
5. **ALMOST EVERY POST expresses some values** - it is rare that ALL 19 values are truly absent. Look for implicit values in tone, word choice, context. Even neutral posts often imply some value (e.g., information sharing shows stimulation/achievement, personal updates show pleasure/caring).

## Likert Scale

Rate each value from 0 to 6:

- **0**: This value is absent or contradicted
- **1-2**: This value is present but subtle/implied
- **3-4**: This value is moderately expressed
- **5-6**: This value is strongly predominant

## Output Format

OUTPUT EXACTLY ONE JSON OBJECT. DO NOT USE MARKDOWN CODE BLOCKS, BACKTICKS, OR ANY FORMATTING. Return ONLY raw JSON.

RETURN A JSON OBJECT WITH EXACTLY TWO FIELDS:

- "Rating": object with 19 key-value pairs (concept name: integer rating 0-6)
- "Reasoning": a SINGLE STRING containing a brief summary of your reasoning

CRITICAL: "Reasoning" MUST be a single string value, NOT an object.

IMPORTANT: Use EXACTLY these JSON keys (note the 's' at the end):
- "independent_thoughts" (NOT "independent_thought")
- "independent_actions" (NOT "independent_action")

Example format:
{"Rating": {"reputation": 0, "power": 6, "wealth": 5, "achievement": 3, "pleasure": 0, "independent_thoughts": 0, "independent_actions": 0, "stimulation": 2, "personal_security": 4, "societal_security": 3, "tradition": 5, "lawfulness": 4, "respect": 0, "humility": 0, "responsibility": 0, "caring": 0, "equality": 0, "nature": 0, "tolerance": 0}, "Reasoning": "This is a single string explaining the overall reasoning."}

## Examples

1. **Political Action Post**
   Post: "I've always believed in the power of research to save lives and ensure Americans get the care they need.
   Starting today, the first-ever White House Initiative on Women's Health Research will work towards that goal, changing how we approach and fund women's health research."

{"Rating": {"reputation": 3, "power": 2, "wealth": 0, "achievement": 3, "pleasure": 0, "independent_thoughts": 3, "independent_actions": 4, "stimulation": 1, "personal_security": 0, "societal_security": 2, "tradition": 0, "lawfulness": 0, "respect": 0, "humility": 0, "responsibility": 5, "caring": 6, "equality": 6, "nature": 0, "tolerance": 2}, "Reasoning": "Strong emphasis on caring (5) and responsibility (4) through healthcare initiatives. Moderate achievement (3) for institutional success. Equality and tolerance reflected in women's health focus."}

1. **War Reporting Post**
   Post: "This little kid was carrying a white flag, and now he's dead. This guy was also carrying a white flag, and he's been shot.
   I'm here filming for you, and I'm in a lot of danger as well. People holding white flags are trying to come out, and are scared from the snipers.
   If the claims of civilians with white flags getting INTENTIONALLY targeted is true, this would be a WAR CRIME."

{"Rating": {"reputation": 0, "power": 0, "wealth": 0, "achievement": 0, "pleasure": 0, "independent_thoughts": 0, "independent_actions": 3, "independent_actions": 3, "stimulation": 0, "personal_security": 3, "societal_security": 4, "tradition": 0, "lawfulness": 6, "respect": 3, "humility": 0, "responsibility": 4, "caring": 6, "equality": 6, "nature": 0, "tolerance": 4}, "Reasoning": "Very strong caring (6) and equality (6) through reporting civilian suffering. Strong lawfulness (5) concern for war crimes. Moderate societal security (4) and responsibility (4) in exposing danger."}

1. **Casual Entertainment Post**
   Post: "Just watched the new Marvel movie with friends. The action scenes were incredible! Can't wait for the sequel!"

{"Rating": {"reputation": 0, "power": 0, "wealth": 0, "achievement": 0, "pleasure": 5, "independent_thoughts": 0, "independent_actions": 0, "stimulation": 6, "personal_security": 0, "societal_security": 0, "tradition": 0, "lawfulness": 0, "respect": 0, "humility": 0, "responsibility": 0, "caring": 2, "equality": 0, "nature": 0, "tolerance": 0}, "Reasoning": "Strong stimulation (6) and pleasure (5) from entertainment. Moderate caring (2) reflected in shared social experience with friends."}

1. **Daily Life Post**
   Post: "Beautiful morning run today! Nothing beats starting the day with exercise and fresh air. Feeling grateful for this moment."

{"Rating": {"reputation": 0, "power": 0, "wealth": 0, "achievement": 2, "pleasure": 4, "independent_thoughts": 0, "independent_actions": 3, "stimulation": 3, "personal_security": 2, "societal_security": 0, "tradition": 0, "lawfulness": 0, "respect": 0, "humility": 1, "responsibility": 2, "caring": 1, "equality": 0, "nature": 3, "tolerance": 0}, "Reasoning": "Moderate pleasure (4) and stimulation (3) from exercise. Independent action (3) in personal routine. Achievement (2) and responsibility (2) for self-improvement. Nature (3) appreciation for fresh air."}

1. **Work Achievement Post**
   Post: "Finally got that promotion I've been working toward for 3 years! Hard work pays off. Thanks to everyone who supported me on this journey."

{"Rating": {"reputation": 3, "power": 2, "wealth": 0, "achievement": 6, "pleasure": 3, "independent_thoughts": 0, "independent_actions": 4, "stimulation": 2, "personal_security": 0, "societal_security": 0, "tradition": 0, "lawfulness": 0, "respect": 0, "humility": 2, "responsibility": 3, "caring": 2, "equality": 0, "nature": 0, "tolerance": 0}, "Reasoning": "Very strong achievement (6) and independent action (4) for career success. Moderate pleasure (3) and responsibility (3). Recognition of supporters reflects caring (2) and reputation (3)."}

