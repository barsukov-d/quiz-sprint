# Screen: Home Page

**Purpose:** Main screen with two game modes: Daily Challenge and Classic Mode.

> Follows `/docs/ui/UI_GUIDELINES.md`

---

## Layout

**Header:**
- Logo "Quiz Sprint" (left)
- User avatar (right)

**Content:**
- Daily Challenge card (top, prominent)
- Classic Mode card (below)

**Navigation:**
- Bottom tab bar: [Home] [Leaderboard] [Profile]

---

## Visual Elements

### Daily Challenge Card

**Style:** Gradient border (orange-red), largest element, top position

**Content:**
- "🔥 Daily Challenge"
- Quiz title: "World Capitals"
- "10 questions"
- "🔥 5 days streak"
- "+50% bonus points"
- "Resets in: 14h 23m"
- Button: `[Play Daily Quiz]`

**States:**
- Available: Button enabled, pulsing glow
- Completed: Green ✅, button disabled, shows "Completed"
- Loading: Skeleton animation

---

### Classic Mode Card

**Style:** Simple card, neutral background

**Content:**
- "🎮 Classic Mode"
- "Choose any quiz and beat your records"
- Button: `[Browse Quizzes]`

**States:**
- Default: Button enabled
- Loading: Skeleton animation

---

### Bottom Navigation

**Style:** Fixed bar, 3 tabs

**Tabs:**
- [Home] - highlighted/active
- [Leaderboard]
- [Profile]

---

## Wireframe

```text
┌────────────────────────────────────────────┐
│  Quiz Sprint                        [👤]  │
├────────────────────────────────────────────┤
│                                            │
│ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓ │
│ ┃  🔥 DAILY CHALLENGE                  ┃ │
│ ┃                                      ┃ │
│ ┃  World Capitals                      ┃ │
│ ┃  10 questions                        ┃ │
│ ┃                                      ┃ │
│ ┃  🔥 5 days streak                    ┃ │
│ ┃  +50% bonus points                   ┃ │
│ ┃  Resets in: 14h 23m                  ┃ │
│ ┃                                      ┃ │
│ ┃       [Play Daily Quiz]              ┃ │
│ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛ │
│                                            │
│ ┌─────────────────────────────────────┐   │
│ │  🎮 CLASSIC MODE                    │   │
│ │                                     │   │
│ │  Choose any quiz and beat your     │   │
│ │  records                            │   │
│ │                                     │   │
│ │        [Browse Quizzes]             │   │
│ │                                     │   │
│ └─────────────────────────────────────┘   │
│                                            │
└────────────────────────────────────────────┘
       [Home] [Leaderboard] [Profile]
```

---

## Visual Hierarchy

1. **Daily Challenge** - Bright gradient, largest, top priority
2. **Classic Mode** - Neutral card, secondary
3. **Navigation** - Fixed, always visible

---

## Interactions

**Daily Challenge:**
- **Tap card or button** → Navigate to game (daily mode)

**Classic Mode:**
- **Tap `[Browse Quizzes]`** → Navigate to quiz list screen
- **Tap card** → Navigate to quiz list screen

**Navigation:**
- **Tap [Home]** → Current screen
- **Tap [Leaderboard]** → Leaderboard screen
- **Tap [Profile]** → Profile screen
- **Tap avatar** → Profile screen

---

## Animations

**Daily Challenge:**
- Not completed: Pulsing glow
- Just completed: Confetti burst
- Streak update: +1 counter animation

**Navigation:**
- Tab switch: Slide transition
