# Screen: Leaderboard

**Purpose:** Show player rankings for specific quiz to enable social comparison and competition.

> Follows `/docs/ui/UI_GUIDELINES.md`

---

## Layout

**Header:**
- Back button (left)
- Quiz name (center)
- User avatar (right)

**Content:**
- Top 3 players (special styling)
- Ranking table (scrollable)
- Current user row (highlighted, sticky)

**Navigation:**
- Bottom tab bar: [Home] [Leaderboard] [Profile]

---

## Visual Elements

### Top 3 Section

**Style:** Prominent area at top, special background

**Content:**
- 1st place: 🥇 avatar, username, score
- 2nd place: 🥈 avatar, username, score
- 3rd place: 🥉 avatar, username, score

---

### Ranking Table

**Style:** List layout, alternating row colors

**Columns:**
- Rank (number)
- Player (avatar + username)
- Score (points)
- Date (relative, e.g., "2d ago")

**Row States:**
- Default: Normal styling
- Current user: Highlighted background, bold text
- Loading: Skeleton rows

---

### Current User Position

**Style:** Sticky row at bottom (if user not in top 50)

**Content:**
- "Your position: #156"
- Avatar, username, score
- Highlighted with accent color

**States:**
- In top 50: Highlighted in main list
- Below top 50: Sticky row at bottom
- Not ranked: "Complete quiz to get ranked"

---

### Empty State

**Style:** Centered message

**Content:**
- Icon: 🏆
- "No rankings yet"
- "Be the first to complete this quiz!"

---

## Wireframe

```text
┌────────────────────────────────────────────┐
│  ←  World Capitals Quiz             [👤]  │
├────────────────────────────────────────────┤
│  ┌────────────────────────────────────┐   │
│  │  TOP 3                             │   │
│  │  🥇 Alice      1250                │   │
│  │  🥈 Bob        1180                │   │
│  │  🥉 Charlie    1050                │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │ #4  Dave        980    1d ago      │   │
│  ├────────────────────────────────────┤   │
│  │ #5  Eve         920    2d ago      │   │
│  ├────────────────────────────────────┤   │
│  │ #6  Frank       890    3d ago      │   │
│  ├────────────────────────────────────┤   │
│  │ #7  YOU         850    5d ago  ⭐  │   │
│  ├────────────────────────────────────┤   │
│  │ #8  Grace       820    1w ago      │   │
│  └────────────────────────────────────┘   │
│           [Load More]                      │
│                                            │
└────────────────────────────────────────────┘
       [Home] [Leaderboard] [Profile]
```

---

## Visual Hierarchy

1. **Top 3** - Most prominent, special medals, larger size
2. **Current user row** - Highlighted, stands out in list
3. **Other players** - Standard list items

---

## Interactions

**Table:**
- **Scroll list** → Load more rankings
- **Pull down** → Refresh rankings
- **Tap player row** → View player profile (future)
- **Reach bottom** → Auto-load next 50 entries

**Navigation:**
- **Tap back button** → Return to previous screen
- **Tap avatar** → Navigate to Profile
- **Tap [Load More]** → Load next page

**Auto-scroll:**
- On open: Scroll to current user position

---

## Animations

**Initial load:**
- Top 3 fade in with scale
- Table rows fade in top-to-bottom

**Pull-to-refresh:**
- Spinner at top while loading

**New rankings:**
- Smooth position transitions
- Current user row pulse highlight

**Empty state:**
- Trophy icon bounce animation
