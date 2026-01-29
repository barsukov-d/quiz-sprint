# Screen: Profile

**Purpose:** Display personal progress, statistics, and achievements.

> Follows `/docs/ui/UI_GUIDELINES.md`

---

## Layout

**Header:**
- Back button (left)
- "Profile" title (center)

**Content:**
- User info card (top)
- Statistics section
- Achievements section
- Recent activity list
- Settings/actions (bottom)

**Navigation:**
- Bottom tab bar: [Home] [Leaderboard] [Profile]

---

## Visual Elements

### User Info Card

**Style:** Card at top, centered content

**Content:**
- Avatar (Telegram photo)
- Username: "@username"
- Display name: "John Doe"
- Member since: "Joined Jan 2024"

---

### Statistics Section

**Style:** Grid or card layout

**Content:**
- "📊 Statistics"
- Quizzes completed: "12"
- Total points: "1,250"
- Average score: "78%"
- Best rank: "#3"
- Time played: "2h 15m"

---

### Achievements Section

**Style:** Grid of badge icons

**Content:**
- "🏆 Achievements"
- Unlocked badges: ✅ icon, colored
- Locked badges: 🔒 icon, grayscale
- Progress bar (for locked badges)

**Badge examples:**
- "First Quiz" - Complete your first quiz
- "Top 10" - Reach top 10 in any leaderboard
- "Speed Demon" - Complete quiz in under 2 minutes
- "Perfect Score" - Get 100% correct answers

---

### Recent Activity

**Style:** List, scrollable

**Content:**
- "📜 Recent Activity"
- Last 5 completed quizzes
- Each entry shows:
  - Quiz name
  - Score
  - Rank: "#12"
  - Time: "2h ago"

---

### Settings Section

**Style:** List of action buttons

**Content:**
- Edit profile
- Language selection
- Notification settings
- Logout

---

## Wireframe

```text
┌────────────────────────────────────────────┐
│  ←  Profile                                │
├────────────────────────────────────────────┤
│         ┌────────────────┐                 │
│         │      👤        │                 │
│         │  @username     │                 │
│         │  John Doe      │                 │
│         │  Joined Jan 24 │                 │
│         └────────────────┘                 │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  📊 Statistics                     │   │
│  │  Quizzes: 12   Points: 1,250      │   │
│  │  Avg: 78%      Best: #3           │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  🏆 Achievements                   │   │
│  │  [✅ First] [✅ Top10] [🔒 Speed]  │   │
│  │  [🔒 Perfect] [🔒 Marathon]        │   │
│  └────────────────────────────────────┘   │
│                                            │
│  ┌────────────────────────────────────┐   │
│  │  📜 Recent Activity                │   │
│  │  • World Capitals  850  #7  2h ago │   │
│  │  • Science Quiz    920  #5  1d ago │   │
│  │  • Sports Trivia   780  #12 2d ago │   │
│  └────────────────────────────────────┘   │
│                                            │
│  [Edit Profile]  [Settings]  [Logout]     │
│                                            │
└────────────────────────────────────────────┘
       [Home] [Leaderboard] [Profile]
```

---

## Visual Hierarchy

1. **User info** - Most prominent, centered, top
2. **Statistics** - Key metrics, clear numbers
3. **Achievements** - Visual badges, gamification
4. **Recent activity** - Supporting info
5. **Settings** - Bottom, less prominent

---

## Interactions

**User Info:**
- **Tap avatar** → Edit profile photo
- **Tap [Edit Profile]** → Edit profile screen

**Statistics:**
- Static display (no interactions)

**Achievements:**
- **Tap locked badge** → Show unlock requirements
- **Tap unlocked badge** → Show achievement details

**Recent Activity:**
- **Tap activity item** → Navigate to quiz leaderboard

**Settings:**
- **Tap [Edit Profile]** → Edit profile form
- **Tap [Settings]** → Settings screen
- **Tap [Logout]** → Logout confirmation dialog

**Navigation:**
- **Tap back button** → Return to previous screen
- **Tap [Profile] tab** → Current screen (no change)

---

## Animations

**Initial load:**
- Avatar fade in with scale
- Sections fade in top-to-bottom

**Achievement unlock:**
- Badge pulse and scale animation
- Confetti effect

**Statistics update:**
- Number count-up animation
