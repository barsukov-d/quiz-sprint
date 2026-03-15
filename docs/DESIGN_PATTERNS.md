# Design Patterns — Quizzo UI Kit

Reference document for styling Quiz Sprint TMA components.
Source: Quizzo UI Kit (220+ screens, iOS/Android, Light & Dark).

---

## 1. Colors

### Primary Scale (`#6C5CE7`)

| Shade | Hex       | Usage                              |
|-------|-----------|------------------------------------|
| 50    | `#f0edff` | Light bg accents, hover states     |
| 100   | `#e0dcfe` | Light borders, subtle fills        |
| 200   | `#c2bafd` | Light mode active indicators       |
| 300   | `#a397fa` | Light mode secondary elements      |
| 400   | `#8675f3` | **Dark theme primary** (`--ui-primary`) |
| 500   | `#6c5ce7` | Brand color, illustrations         |
| 600   | `#5847d0` | **Light theme primary** (`--ui-primary`) |
| 700   | `#4536b2` | Hover/pressed button states        |
| 800   | `#352a8f` | Deep accents                       |
| 900   | `#261e6c` | Very dark purple                   |
| 950   | `#1a144a` | Darkest purple                     |

### Dark Theme Backgrounds

| Token              | Hex       | Usage                     |
|---------------------|-----------|---------------------------|
| `--ui-bg`           | `#12141e` | Page background           |
| `--ui-bg-muted`     | `#181a28` | Sections, groupings       |
| `--ui-bg-elevated`  | `#1f2236` | Cards, modals, sheets     |
| `--ui-bg-accented`  | `#262a40` | Inputs, highlighted areas |

### Dark Theme Borders

| Token                  | Hex       |
|------------------------|-----------|
| `--ui-border`          | `#2e3248` |
| `--ui-border-muted`    | `#262a40` |
| `--ui-border-accented` | `#3a3e56` |

### Light Theme Backgrounds

| Token              | Hex/Value  | Usage                          |
|---------------------|------------|--------------------------------|
| `--ui-bg`           | `#ffffff`  | Page background                |
| `--ui-bg-muted`     | `#f8f7fc`  | Sections (subtle purple tint)  |
| `--ui-bg-elevated`  | `#ffffff`  | Cards, modals                  |
| `--ui-bg-accented`  | `#f0eeff`  | Inputs, highlighted areas      |

### Semantic Colors

| Role    | Color Name | Hex (approx) | Usage                         |
|---------|-----------|--------------|-------------------------------|
| Success | emerald   | `#27AE60`    | Correct answer, positive      |
| Error   | rose      | `#EB5757`    | Incorrect answer, destructive |
| Warning | amber     | `#F2994A`    | Close/near miss, caution      |
| Info    | sky       | `#2F80ED`    | Informational, links          |

### Quiz Answer Option Colors

4 answer buttons displayed in 2x2 grid during gameplay. Each has a distinct color:

| Position    | Color   | Hex (approx) | Tailwind             |
|-------------|---------|--------------|----------------------|
| Top-left    | Red     | `#E74C3C`    | `bg-red-500`         |
| Top-right   | Green   | `#2ECC71`    | `bg-emerald-500`     |
| Bottom-left | Yellow  | `#F1C40F`    | `bg-yellow-500`      |
| Bottom-right| Blue    | `#3498DB`    | `bg-sky-500`         |

In dark theme — same colors, slightly adjusted saturation for readability.

---

## 2. Typography

### Font Family

Design uses a geometric sans-serif (Urbanist/similar). Project uses **Inter**.

```
font-family: 'Inter', ui-sans-serif, system-ui, -apple-system, sans-serif;
```

### Scale

| Role               | Size   | Weight     | Tailwind                  | Usage                            |
|--------------------|--------|------------|---------------------------|----------------------------------|
| Page title         | 24px   | Bold (700) | `text-2xl font-bold`      | "Quizzo", "Discover", "Profile"  |
| Section heading    | 20px   | SemiBold   | `text-xl font-semibold`   | "Top Authors", "Discover"        |
| Card title         | 16px   | SemiBold   | `text-base font-semibold` | Quiz name, collection name       |
| Body               | 14px   | Regular    | `text-sm`                 | Descriptions, paragraphs         |
| Caption / metadata | 12px   | Regular    | `text-xs`                 | "3 months ago", "4.9K plays"     |
| Tiny label         | 10px   | Medium     | `text-[10px] font-medium` | Badge text, small counters       |
| Quiz question      | 18px   | SemiBold   | `text-lg font-semibold`   | Question text (centered)         |
| Score overlay      | 28px   | Bold       | `text-3xl font-bold`      | "+945", "+2548" score badge      |
| Scoreboard name    | 14px   | Medium     | `text-sm font-medium`     | Player names in leaderboard      |
| Stat numbers       | 20px   | Bold       | `text-xl font-bold`       | "265", "32M", "27.4M"           |

### Text Colors

| Role        | Light Theme       | Dark Theme        | Tailwind                  |
|-------------|-------------------|-------------------|---------------------------|
| Primary     | neutral-700       | neutral-300       | `text-default` (auto)     |
| Highlighted | neutral-950       | neutral-50        | `text-highlighted`        |
| Muted       | neutral-500       | neutral-400       | `text-muted`              |
| Dimmed      | neutral-400       | neutral-500       | `text-dimmed`             |
| On primary  | white             | white             | `text-white`              |
| Link        | primary-600       | primary-400       | `text-primary`            |

---

## 3. Border Radius

| Element                | Radius   | Tailwind        | Notes                         |
|------------------------|----------|-----------------|-------------------------------|
| CTA Buttons            | full     | `rounded-full`  | Pill-shaped, always           |
| "Follow" buttons       | full     | `rounded-full`  | Small pill                    |
| Tab pills / segments   | full     | `rounded-full`  | Active/inactive tabs          |
| Tags / chips           | full     | `rounded-full`  | Category labels               |
| Cards                  | 16px     | `rounded-2xl`   | Quiz cards, profile cards     |
| Modals / sheets        | 20px top | `rounded-t-[20px]` | Bottom sheets             |
| Input fields           | 12px     | `rounded-xl`    | Text inputs, selects          |
| Answer option buttons  | 12-16px  | `rounded-xl`    | Quiz gameplay 2x2 grid        |
| Avatar                 | full     | `rounded-full`  | Always circular               |
| Score badge            | 8px      | `rounded-lg`    | "+945" floating badge         |
| Progress bar           | full     | `rounded-full`  | Thin bar, fully rounded       |
| Bottom nav bar         | 0        | `rounded-none`  | Flat bottom edge              |
| Image thumbnails       | 12px     | `rounded-xl`    | Quiz cover images             |
| Account type cards     | 16px     | `rounded-2xl`   | Onboarding selection cards    |
| Collection grid items  | 12px     | `rounded-xl`    | Image grid in collections     |

**Global `--ui-radius`:** `0.75rem` (12px) — most common base value in design.

---

## 4. Shadows

| Context       | Light Theme                          | Dark Theme    |
|---------------|--------------------------------------|---------------|
| Cards         | `shadow-sm` (subtle, gray)           | None (bg contrast only) |
| Elevated card | `shadow-md`                          | None          |
| Buttons       | None or very subtle                  | None          |
| Modals        | `shadow-xl`                          | None          |
| Bottom nav    | `shadow-[0_-1px_3px_rgba(0,0,0,0.1)]` | None        |

Dark theme relies entirely on background color layering, not shadows.

---

## 5. Buttons

### Primary CTA

```
Full-width, pill-shaped, solid primary bg, white text
Height: ~48px (h-12)
```

| Property   | Value                                       |
|------------|---------------------------------------------|
| Shape      | `rounded-full w-full h-12`                  |
| Background | `bg-primary` (purple)                       |
| Text       | `text-white font-semibold text-base`        |
| Examples   | "GET STARTED", "SIGN IN", "Continue"        |

### Secondary / Ghost CTA

```
Full-width, pill-shaped, transparent bg, border or text-only
```

| Property   | Value                                       |
|------------|---------------------------------------------|
| Shape      | `rounded-full w-full h-12`                  |
| Background | transparent                                 |
| Border     | `border border-primary` or none             |
| Text       | `text-primary font-semibold`                |
| Examples   | "I ALREADY HAVE AN ACCOUNT", "Skip"         |

### Small Action Button

```
Inline, pill-shaped, solid primary
```

| Property   | Value                                       |
|------------|---------------------------------------------|
| Shape      | `rounded-full px-4 py-1.5`                  |
| Background | `bg-primary`                                |
| Text       | `text-white text-xs font-medium`            |
| Examples   | "Follow", "View all"                        |

### Tab / Segment Buttons

```
Horizontal group, pill-shaped, one active
```

| State    | Style                                        |
|----------|----------------------------------------------|
| Active   | `bg-primary text-white rounded-full px-4 py-2` |
| Inactive | `bg-transparent text-muted rounded-full px-4 py-2` |
| Examples | "Quiz | People | Collections", "My Quizzes | Favorites | Collaboration" |

### Quiz Answer Buttons

```
Large colored blocks in 2x2 grid
```

| Property   | Value                                       |
|------------|---------------------------------------------|
| Shape      | `rounded-xl` (~12-16px)                     |
| Size       | Equal width/height, fills half of grid      |
| Layout     | `grid grid-cols-2 gap-3`                    |
| Text       | `text-white font-semibold text-center`      |
| Colors     | Red, Green, Yellow, Blue (see answer colors) |

### Dual Action Buttons (side by side)

```
Two buttons, equal width, one primary one outline
```

| Property   | Value                                       |
|------------|---------------------------------------------|
| Layout     | `flex gap-3`                                |
| Primary    | `flex-1 rounded-full bg-primary text-white` |
| Secondary  | `flex-1 rounded-full border border-primary text-primary` |
| Examples   | "Play Solo" + "Play with Friends"           |

---

## 6. Cards

### Quiz Card (Horizontal — list item)

```
┌──────────────────────────────────┐
│ [thumb]  Title                   │
│          3 months ago · 4.9K pl  │
│          [avatar] Author Name    │
└──────────────────────────────────┘
```

| Property    | Value                                    |
|-------------|------------------------------------------|
| Shape       | `rounded-2xl`                            |
| Background  | `bg-elevated` (dark: `#1f2236`, light: white) |
| Padding     | `p-3` or `p-4`                           |
| Thumbnail   | `rounded-xl w-16 h-16 object-cover`      |
| Layout      | `flex gap-3 items-start`                 |

### Quiz Card (Vertical — carousel/grid)

```
┌─────────────┐
│  [image]    │
│  Title      │
│  metadata   │
└─────────────┘
```

| Property    | Value                                    |
|-------------|------------------------------------------|
| Shape       | `rounded-2xl`                            |
| Image       | `rounded-t-2xl w-full aspect-video`      |
| Padding     | Body: `p-3`                              |
| Width       | ~160px in carousel, full in grid         |

### Profile / Stats Card

```
┌──────────────────────────────────┐
│        [large avatar]            │
│        Username                  │
│   265    32M     27.4M           │
│   Quizzes  Play   Players       │
│  [Quiz] [Collections] [About]   │
└──────────────────────────────────┘
```

| Property    | Value                                    |
|-------------|------------------------------------------|
| Stats       | `text-xl font-bold` number + `text-xs text-muted` label |
| Tabs below  | Segment buttons (pill tabs)              |

### Onboarding Selection Card

```
┌─────────────────────┐
│ [icon]  Label       │
└─────────────────────┘
```

| Property    | Value                                    |
|-------------|------------------------------------------|
| Shape       | `rounded-2xl`                            |
| Background  | `bg-elevated`                            |
| Layout      | `flex items-center gap-3 p-4`            |
| Icon        | Colored square icon `rounded-xl w-10 h-10` |
| Colors      | Each option has a unique bg color (red, blue, green, orange) |

---

## 7. Navigation

### Top Bar

```
[←]          Page Title          [🔍] [⚙️]
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Height       | ~44px (`h-11`)                          |
| Padding      | `px-4`                                  |
| Back button  | Icon only, left-aligned                 |
| Title        | `text-lg font-semibold` centered        |
| Actions      | Icon buttons, right-aligned             |
| Background   | Transparent (inherits page bg)          |

### Bottom Tab Bar

```
┌─────────────────────────────────────┐
│  🏠    🔍    ➕    📚    👤       │
│ Home  Search Create Library Profile │
└─────────────────────────────────────┘
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Height       | ~64px (`h-16`) + safe area              |
| Items        | 5 icons with labels                     |
| Active       | Primary color icon + label              |
| Inactive     | `text-muted` icon + label               |
| Background   | Dark: `bg-muted`, Light: white          |
| Create btn   | May be larger/elevated (+ icon)         |
| Border top   | Dark: none, Light: `border-t border-muted` |

### Horizontal Tabs / Segments

```
[Quiz]  [People]  [Collections]
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Container    | `flex gap-2 overflow-x-auto`            |
| Active item  | `bg-primary text-white rounded-full px-4 py-2` |
| Inactive     | `text-muted rounded-full px-4 py-2`    |
| Scrollable   | Yes, horizontal scroll on overflow      |

---

## 8. Forms / Inputs

### Text Input

| Property     | Light                    | Dark                     |
|--------------|--------------------------|--------------------------|
| Background   | white                    | `#262a40` (bg-accented)  |
| Border       | `border border-muted`    | None (bg contrast only)  |
| Radius       | `rounded-xl` (12px)      | `rounded-xl`             |
| Height       | ~48px (`h-12`)           | ~48px                    |
| Padding      | `px-4`                   | `px-4`                   |
| Text         | `text-sm`                | `text-sm`                |
| Placeholder  | `text-dimmed`            | `text-dimmed`            |
| Label        | `text-sm font-medium mb-1.5` above field |               |

### Select / Dropdown

Same as text input + chevron icon right-aligned.

### Password Input

Same as text input + eye toggle icon right-aligned.

### Checkbox

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Shape        | `rounded-md` (~4-6px)                   |
| Size         | 20x20px                                 |
| Checked      | `bg-primary` + white checkmark          |
| Unchecked    | `border border-muted`                   |

### PIN / OTP Input

```
[4] [6] [7] [_]
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Cell shape   | `rounded-xl` square ~48x48px            |
| Background   | `bg-elevated` (dark), white (light)     |
| Border       | Active: `border-primary`, inactive: `border-muted` |
| Text         | `text-2xl font-bold` centered           |

---

## 9. Quiz Gameplay

### Question Screen Layout

```
┌──────────────────────────────┐
│ 1/10   Quiz    ⏱            │  ← counter + timer
│ ████████████░░░░░░░░░░░░░░░ │  ← progress bar
│                              │
│        [Question Image]      │  ← optional image
│                              │
│   ..... do you get to        │  ← question text
│   school? by bus?            │
│                              │
│  ┌──────┐  ┌──────┐         │  ← 2x2 answer grid
│  │  How │  │ What │         │
│  └──────┘  └──────┘         │
│  ┌──────┐  ┌──────┐         │
│  │ Which│  │ Where│         │
│  └──────┘  └──────┘         │
│                              │
│  [         Next          ]   │  ← CTA button
└──────────────────────────────┘
```

### Progress Bar

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Height       | ~4px                                    |
| Track        | `bg-muted rounded-full`                 |
| Fill         | Gradient or solid color                 |
| Colors       | Changes per question type (yellow, blue, green) |
| Tailwind     | `h-1 rounded-full`                      |

### Question Counter

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Format       | "1/10"                                  |
| Style        | `text-sm font-semibold`                 |
| Position     | Top-left                                |

### Answer Feedback Overlay

**Correct:**
```
┌──────────────────────────────┐
│        Correct!              │  green bg header
│         +945                 │  score badge
│      [question image]        │
│      question text           │
│      [highlighted answer]    │
│      [      Next      ]      │
└──────────────────────────────┘
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Header bg    | `bg-emerald-500` / green                |
| Badge        | `bg-emerald-600 text-white rounded-lg px-3 py-1` |
| Badge text   | `+945` in bold                          |

**Incorrect:**
```
Same layout, red header
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Header bg    | `bg-red-500`                            |
| Badge text   | "Fuuhhh. That was close" etc.           |

### Scoreboard (Final)

```
        [2nd]  [1st]  [3rd]       ← podium with avatars
         Pedro Andrew  Freida
         3,645  3,645  3,170

   4   [avatar] Clinton    2,846
   5   [avatar] Theresa    2,472
   ...
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Podium       | 3 columns, center elevated              |
| Top 3 avatars| Larger, with medal/crown decoration     |
| Score badge  | `bg-primary rounded-full px-2 py-0.5 text-xs` |
| List row     | `flex items-center gap-3 py-3`          |
| Rank number  | `text-sm font-semibold w-6`             |

---

## 10. Avatars

| Size     | Pixels | Tailwind    | Usage                      |
|----------|--------|-------------|----------------------------|
| XS       | 24px   | `size-6`    | Inline mentions             |
| SM       | 32px   | `size-8`    | List items, comments        |
| MD       | 40px   | `size-10`   | Author rows, quiz cards     |
| LG       | 56px   | `size-14`   | Profile header              |
| XL       | 80px   | `size-20`   | Profile page, scoreboard    |

All avatars: `rounded-full`, always circular.

### Avatar Group (overlap)

```
[😀][😎][🤓] +5
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Overlap      | `-space-x-2`                            |
| Border       | `ring-2 ring-bg` (matches page bg)      |
| Counter      | `bg-muted text-xs rounded-full`         |

---

## 11. Spacing System

| Token      | Value   | Usage                              |
|------------|---------|------------------------------------|
| Page px    | 16px    | `px-4` — horizontal page padding   |
| Section gap| 24px    | `gap-6` — between sections         |
| Card padding| 16px   | `p-4` — inside cards               |
| Card gap   | 12px    | `gap-3` — between cards in list    |
| Inline gap | 8px     | `gap-2` — between inline elements  |
| Input gap  | 16px    | `space-y-4` — between form fields  |
| Label gap  | 6px     | `mb-1.5` — label to input          |
| Bottom safe| 64px+   | Bottom nav height + safe area      |

---

## 12. Modals & Sheets

### Bottom Sheet

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Radius       | `rounded-t-[20px]`                      |
| Background   | `bg-elevated`                           |
| Padding      | `p-4` or `p-6`                          |
| Handle       | `w-10 h-1 bg-muted rounded-full mx-auto mb-4` |

### Centered Modal (e.g., success states)

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Radius       | `rounded-2xl`                           |
| Background   | `bg-elevated`                           |
| Padding      | `p-6`                                   |
| Content      | Illustration + text + CTA button        |
| Max width    | ~320px                                  |

### Selection Grid (Add Question, Time Limit, Points)

```
┌─────┐ ┌─────┐
│ Quiz│ │ T/F │
├─────┤ ├─────┤
│Puzzl│ │Type │
└─────┘ └─────┘
```

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Layout       | `grid grid-cols-2 gap-3`                |
| Item shape   | `rounded-2xl`                           |
| Item bg      | Each has unique color                   |
| Item content | Icon + label, centered                  |

---

## 13. Lists

### Quiz List Item

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Layout       | `flex gap-3 items-start p-3`            |
| Thumbnail    | `rounded-xl w-16 h-16 shrink-0`        |
| Title        | `text-sm font-semibold line-clamp-2`    |
| Metadata     | `text-xs text-muted`                    |
| Divider      | None (card bg separation)               |

### Author / User Row

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Layout       | `flex items-center gap-3 py-2`          |
| Avatar       | `size-10 rounded-full`                  |
| Name         | `text-sm font-semibold`                 |
| Username     | `text-xs text-muted`                    |
| Action       | "Follow" pill button, right-aligned     |

### Settings List

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Layout       | `flex items-center justify-between py-3`|
| Icon         | Left icon `size-5 text-muted`           |
| Label        | `text-sm`                               |
| Chevron      | Right `>` icon                          |
| Divider      | `border-b border-muted`                 |

---

## 14. Onboarding

### Step Screen

```
┌──────────────────────────────┐
│        [illustration]        │
│                              │
│     Title text here          │
│     Subtitle description     │
│                              │
│     ●  ○  ○                  │  ← dot indicators
│                              │
│  [    GET STARTED        ]   │  ← primary CTA
│  [I ALREADY HAVE AN ACCOUNT] │  ← ghost CTA
└──────────────────────────────┘
```

### Dot Indicators

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Active       | `w-6 h-2 bg-primary rounded-full`       |
| Inactive     | `w-2 h-2 bg-muted rounded-full`         |
| Gap          | `gap-1.5`                               |

### Progress Steps (account setup)

| Property     | Value                                   |
|--------------|-----------------------------------------|
| Track        | `h-1 bg-muted rounded-full`             |
| Fill         | `bg-primary` proportional to step       |

---

## 15. Nuxt UI Component Mapping

Reference for Step 2 — which Nuxt UI components to customize:

| Design Pattern         | Nuxt UI Component   | Key Overrides                    |
|------------------------|---------------------|----------------------------------|
| Primary CTA            | `UButton`           | `rounded-full`, size `xl`        |
| Ghost CTA              | `UButton` ghost     | `rounded-full`, variant `ghost`  |
| Follow button          | `UButton`           | size `xs`, `rounded-full`        |
| Tab segments           | `UTabs`             | Pill variant, `rounded-full`     |
| Cards                  | `UCard`             | `rounded-2xl`, padding           |
| Text input             | `UInput`            | `rounded-xl`, height             |
| Select                 | `USelect`           | `rounded-xl`                     |
| Checkbox               | `UCheckbox`         | Primary color                    |
| PIN input              | `UPinInput`         | `rounded-xl`, size               |
| Avatar                 | `UAvatar`           | Size scale                       |
| Avatar group           | `UAvatarGroup`      | Overlap spacing                  |
| Modal                  | `UModal`            | `rounded-2xl`                    |
| Bottom sheet           | `UDrawer`           | Bottom position                  |
| Badge                  | `UBadge`            | `rounded-full`                   |
| Progress bar           | `UProgress`         | Height, colors                   |
| Separator              | `USeparator`        | Border colors                    |
| Toast                  | `UToast`            | `rounded-xl`                     |
| Navigation menu        | `UNavigationMenu`   | Bottom bar style                 |
| Dropdown               | `UDropdownMenu`     | `rounded-xl`                     |

---

## 16. Global Design Tokens Summary

```css
:root {
  --ui-radius: 0.75rem;          /* 12px — base radius */
}
```

| Token            | Value      | Notes                    |
|------------------|------------|--------------------------|
| `--ui-radius`    | `0.75rem`  | Base for all components  |
| Page padding     | `1rem`     | 16px horizontal          |
| Card padding     | `1rem`     | 16px inner               |
| Section gap      | `1.5rem`   | 24px between blocks      |
| Font             | Inter      | Geometric sans-serif     |
| Base text size   | 14px       | `text-sm`                |
