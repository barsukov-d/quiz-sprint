# Screen: Home Page

**Purpose:** Main screen with two game modes: Daily Challenge and Classic Mode.

---

## Layout

### Header
- App logo "Quiz Sprint" (left)
- User avatar (right, clickable → Profile)

### Content Area

#### 1. Daily Challenge Card (Top)
**Visual Style:** Prominent card with gradient border, stands out

**Content:**
- Header: "🔥 Daily Challenge"
- Quiz title (e.g., "World Capitals")
- Question count (e.g., "10 questions")
- Streak display: "🔥 5 days streak"
- Bonus badge: "+50% bonus points"
- Timer: "Resets in: 14h 23m"
- Primary button: `[Play Daily Quiz]`

**States:**
- **Available:** Button enabled, pulsing animation
- **Completed:** Green checkmark, button shows "Completed", disabled
- **Loading:** Skeleton animation

---

#### 2. Classic Mode Section (Below Daily)
**Visual Style:** Clean section with list layout

**Header:**
- Title: "🎮 Classic Mode"
- Subtitle: "Play any quiz, beat your records"

**Quiz List:**
Each quiz card shows:
- Quiz icon/emoji (left)
- Quiz title (bold)
- Metadata: "X questions • Category"
- Play button (right): `[Play →]`

**Scrolling:** Vertical scroll for long lists

---

### Bottom Navigation
Fixed tab bar with 3 items:
- **[Home]** - active/highlighted
- **[Leaderboard]**
- **[Profile]**

---

## Wireframe

```text
┌────────────────────────────────────────────┐
│  Quiz Sprint                        [👤]   │
├────────────────────────────────────────────┤
│                                            │
│ ┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓   │
│ ┃  🔥 DAILY CHALLENGE                  ┃   │
│ ┃                                      ┃   │
│ ┃  World Capitals                      ┃   │
│ ┃  10 questions                        ┃   │
│ ┃                                      ┃   │
│ ┃  🔥 5 days streak                    ┃   │
│ ┃  +50% bonus points                   ┃   │
│ ┃  Resets in: 14h 23m                  ┃   │
│ ┃                                      ┃   │
│ ┃       [Play Daily Quiz]              ┃   │
│ ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛   │
│                                            │
│ ┌─────────────────────────────────────┐    │
│ │  🎮 CLASSIC MODE                    │    │
│ │  Play any quiz, beat your records   │    │
│ └─────────────────────────────────────┘    │
│                                            │
└────────────────────────────────────────────┘
       [Home] [Leaderboard] [Profile]
```

---

## Visual Hierarchy

### Priority Order:
1. **Daily Challenge** - Largest, most prominent, top position
2. **Classic Mode** - Secondary, scrollable content below

### Color/Style:
- Daily Challenge: Bright gradient, eye-catching
- Classic Mode: Neutral background, clean
- Active elements: High contrast buttons
- Completed state: Muted/disabled appearance

---

## Interactions

### Daily Challenge Card:
- **Tap `[Play Daily Quiz]`** → Navigate to game screen (daily mode)
- **Tap anywhere on card** → Same as button (if available)

### Classic Mode:
- **Tap quiz card** → Navigate to game screen (classic mode)
- **Tap `[Play →]` button** → Navigate to game screen (classic mode)
- **Scroll list** → View more quizzes

### Navigation Bar:
- **Tap [Home]** → Stay on current screen
- **Tap [Leaderboard]** → Navigate to Leaderboard
- **Tap [Profile]** → Navigate to Profile
- **Tap avatar (header)** → Navigate to Profile

---

## Animations

### Daily Challenge:
- **Not completed:** Subtle pulse/glow animation
- **Just completed:** Confetti effect (once)
- **Streak increment:** Counter animates +1

### Classic Mode:
- **Quiz list load:** Fade in from top to bottom

### Navigation:
- **Tab switch:** Slide animation between screens
