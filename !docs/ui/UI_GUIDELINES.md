# UI Documentation Guidelines

**Purpose:** Document WHAT users see and HOW they interact, NOT technical implementation.

---

## Document Structure

```markdown
# Screen: [Name]
**Purpose:** [1 sentence]

## Layout
- Header: [elements]
- Content: [sections]
- Navigation: [items]

## Visual Elements
### [Component]
- **Style:** [appearance]
- **Content:** [text/icons]
- **States:** [variations]

## Wireframe
[ASCII diagram]

## Interactions
- **[Action]** → [Result]

## Animations (optional)
- [Visual feedback]
```

---

## Include ✅

**Visual:**
- Position (top/middle/bottom, left/right/center)
- Size (large/small, prominent/subtle)
- Style (colors, borders, shadows, gradients)
- Icons, emojis, images

**Content:**
- Exact text labels
- Button text
- Example data

**States:**
- Default, active, disabled, loading, error, empty

**Interactions:**
- Tap/click → result
- Scroll, swipe behavior
- Navigation flow

**Feedback:**
- Animations, transitions

---

## Exclude ❌

- API endpoints
- Code/implementation
- Backend logic
- Database schema
- Technical architecture
- Performance details

---

## Writing Style

**Use descriptive visual terms:**
- ✅ "Gradient border, bright colors"
- ❌ "border: linear-gradient(#ff0000, #0000ff)"

**Be specific:**
- ✅ "Shows '🔥 5 days streak'"
- ❌ "Shows streak counter"

**Clear actions:**
- ✅ "Tap [Play] → Navigate to game screen"
- ❌ "onClick navigates to /game"

---

## Review Checklist

- [ ] No APIs or code
- [ ] Clear wireframe
- [ ] All interactions documented
- [ ] States described
- [ ] Visual hierarchy clear
- [ ] Non-technical language
