# Specification: Head-to-Head Duel - Fighting Game Mode
**Context:** DuelGame
**Status:** Draft

## 1. Business Goal (User Story)
> As a player, I want to challenge other players to real-time 1v1 quiz battles, where each correct answer deals damage to my opponent, creating the intensity of a fighting game combined with intellectual competition.

## 2. Ubiquitous Language
- **Duel:** Real-time 1v1 quiz battle with fighting game mechanics.
- **HP (Health Points):** Continuous health system (0-100 per player). First to 0 HP loses.
- **Combo:** Chain of consecutive correct answers that amplifies damage dealt to opponent.
- **Damage:** HP reduction dealt to opponent on correct answer (base: 15 HP).
- **Critical Hit:** Damage bonus for answering within the first 3 seconds after question display.
- **Speed Window:** First 3 seconds of a question when a critical hit can be landed.
- **Special Move:** Auto-triggered powerful attack at Combo milestones (3, 5, 7).
- **Defense (Block):** Damage reduction (50%) when opponent answers faster than you.
- **KO (Knockout):** Victory when opponent HP reaches 0.

## 3. Business Rules & Logic
1. **Base Damage:** Correct answers deal 15 HP damage. Incorrect answers cause 10 HP self-damage and reset Combo. Timeouts cause 5 HP self-damage.
2. **Combo Multipliers:** Damage is multiplied by x1.3 (Combo 2-3), x1.6 (Combo 4-5), and x2.0 (Combo 6+).
3. **Critical Hit:** Answering within the first 3 seconds adds a +50% damage bonus.
4. **Special Moves:** Automatically trigger at Combo 3 (40 HP), Combo 5 (50 HP + 2s stun), and Combo 7 (70 HP). These replace normal damage and reset Combo.
5. **Defense Mechanism:** If both players answer correctly, the faster player deals full damage, while the slower player deals only 50% damage (blocked).
6. **Damage Calculation:** Multipliers are applied sequentially with rounding down after each step: Base Damage → Combo Multiplier → Crit/Block.
7. **Victory Conditions:** A match ends immediately if a player's HP reaches 0 (KO). If all 10 questions are answered, the player with more HP wins.
8. **Matchmaking:** Players are matched based on Rating ±150 points.

## 4. Scenarios (User Flows)
- **Scenario: Critical Combo**
    - **Given:** Both players have Combo = 2, both answer correctly.
    - **When:** Player answers in 2.5s (critical hit), opponent in 5s (slower).
    - **Then:** Player deals 28 HP (15 base → 19 with combo x1.3 → 28 with crit +50%), opponent deals 9 HP (15 base → 19 with combo x1.3 → 9 with block -50%).

- **Scenario: Special Move - Ultimate Attack**
    - **Given:** Player has Combo = 6, opponent has 80 HP.
    - **When:** Player answers the next question correctly.
    - **Then:** Combo reaches 7, triggering "Ultimate Attack" for 70 HP damage; player's Combo resets to 0.

- **Scenario: Perfect KO Victory**
    - **Given:** Player has 100 HP, opponent has 25 HP.
    - **When:** Player answers correctly (Combo 5) and opponent makes a mistake.
    - **Then:** Opponent HP reaches 0, and the match ends immediately with a KO victory.
