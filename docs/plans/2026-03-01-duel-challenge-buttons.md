# Duel Challenge Action Buttons Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add action buttons to outgoing (Share Again) and incoming (block-style Accept/Decline with challenger name) challenge cards in the Duel Lobby.

**Architecture:** Backend adds `challengerUsername` to `ChallengeDTO` via a `userRepo` lookup in `GetDuelStatusUseCase`. Frontend updates `DuelLobbyView.vue` after regenerating TypeScript types from updated Swagger. No new API endpoints.

**Tech Stack:** Go (backend DTO + mapper + use case), swaggo/swag (Swagger), kubb (TS codegen), Vue 3 + i18n (frontend).

---

### Task 1: Add `challengerUsername` to backend DTO

**Files:**
- Modify: `backend/internal/application/quick_duel/dto.go:81-92`

**Step 1: Add the field**

In `ChallengeDTO` struct, add after `ChallengedID`:

```go
type ChallengeDTO struct {
	ID                 string  `json:"id"`
	ChallengerID       string  `json:"challengerId"`
	ChallengedID       *string `json:"challengedId,omitempty"`
	ChallengerUsername string  `json:"challengerUsername,omitempty"`
	Type               string  `json:"type"`
	Status             string  `json:"status"`
	ChallengeLink      string  `json:"challengeLink,omitempty"`
	ExpiresAt          int64   `json:"expiresAt"`
	ExpiresIn          int     `json:"expiresIn"`
	CreatedAt          int64   `json:"createdAt"`
}
```

**Step 2: Verify it compiles**

```bash
cd backend && go build ./...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add backend/internal/application/quick_duel/dto.go
git commit -m "feat(pvp-duel): add challengerUsername to ChallengeDTO"
```

---

### Task 2: Update `ToChallengeDTO` mapper to accept username

**Files:**
- Modify: `backend/internal/application/quick_duel/mapper.go:97-121`

**Step 1: Change function signature and populate field**

```go
// ToChallengeDTO converts domain DuelChallenge to DTO
func ToChallengeDTO(challenge *quick_duel.DuelChallenge, now int64, challengerUsername string) ChallengeDTO {
	var challengedID *string
	if challenge.ChallengedID() != nil {
		id := challenge.ChallengedID().String()
		challengedID = &id
	}

	expiresIn := int(challenge.ExpiresAt() - now)
	if expiresIn < 0 {
		expiresIn = 0
	}

	return ChallengeDTO{
		ID:                 challenge.ID().String(),
		ChallengerID:       challenge.ChallengerID().String(),
		ChallengedID:       challengedID,
		ChallengerUsername: challengerUsername,
		Type:               string(challenge.Type()),
		Status:             string(challenge.Status()),
		ChallengeLink:      challenge.ChallengeLink(),
		ExpiresAt:          challenge.ExpiresAt(),
		ExpiresIn:          expiresIn,
		CreatedAt:          challenge.CreatedAt(),
	}
}
```

**Step 2: Fix compilation errors** — `use_cases.go` calls `ToChallengeDTO` without the new param. Temporarily fix by passing empty string:

In `backend/internal/application/quick_duel/use_cases.go`, change both loops:

```go
// For pendingChallenges loop (~line 76-78):
for _, c := range pendingChallenges {
    challengeDTOs = append(challengeDTOs, ToChallengeDTO(c, now, ""))
}

// For outgoingChallenges loop (~line 88-91):
for _, c := range outgoingChallenges {
    outgoingDTOs = append(outgoingDTOs, ToChallengeDTO(c, now, ""))
}
```

**Step 3: Verify it compiles**

```bash
cd backend && go build ./...
```
Expected: no errors.

**Step 4: Run tests**

```bash
cd backend && go test ./internal/application/quick_duel/...
```
Expected: all pass.

**Step 5: Commit**

```bash
git add backend/internal/application/quick_duel/mapper.go \
        backend/internal/application/quick_duel/use_cases.go
git commit -m "feat(pvp-duel): update ToChallengeDTO to accept challengerUsername"
```

---

### Task 3: Look up challenger username in `GetDuelStatusUseCase`

**Files:**
- Modify: `backend/internal/application/quick_duel/use_cases.go:75-91`

**Step 1: Replace the pendingChallenges loop with username lookup**

Replace the existing loop (lines ~75-78) with:

```go
challengeDTOs := make([]ChallengeDTO, 0, len(pendingChallenges))
for _, c := range pendingChallenges {
    username := c.ChallengerID().String() // fallback = ID
    if u, err := uc.userRepo.FindByID(c.ChallengerID()); err == nil && u != nil {
        if u.Username().String() != "" {
            username = u.Username().String()
        } else if u.TelegramUsername().String() != "" {
            username = "@" + u.TelegramUsername().String()
        }
    }
    challengeDTOs = append(challengeDTOs, ToChallengeDTO(c, now, username))
}
```

Outgoing challenges: keep passing empty string (current user is the challenger, no need to look up self).

**Step 2: Verify it compiles**

```bash
cd backend && go build ./...
```

**Step 3: Run tests**

```bash
cd backend && go test ./internal/application/quick_duel/...
```
Expected: all pass (existing tests don't check `ChallengerUsername`).

**Step 4: Commit**

```bash
git add backend/internal/application/quick_duel/use_cases.go
git commit -m "feat(pvp-duel): populate challengerUsername in GetDuelStatus response"
```

---

### Task 4: Update Swagger `DuelChallengeDTO` and regenerate TypeScript

**Files:**
- Modify: `backend/internal/infrastructure/http/handlers/swagger_models.go:1349-1360`
- Run: `pnpm run generate:all` in `tma/`

**Step 1: Add field to `DuelChallengeDTO`**

```go
// DuelChallengeDTO represents a pending challenge
type DuelChallengeDTO struct {
	ID                 string  `json:"id"`
	ChallengerID       string  `json:"challengerId"`
	ChallengedID       *string `json:"challengedId,omitempty"`
	ChallengerUsername string  `json:"challengerUsername,omitempty"`
	Type               string  `json:"type"`
	Status             string  `json:"status"`
	ChallengeLink      string  `json:"challengeLink,omitempty"`
	ExpiresAt          int64   `json:"expiresAt"`
	ExpiresIn          int     `json:"expiresIn"`
	CreatedAt          int64   `json:"createdAt"`
}
```

**Step 2: Regenerate Swagger + TypeScript**

```bash
cd tma && pnpm run generate:all
```
Expected: no errors. Check that `tma/src/api/generated/types/internalInfrastructureHttpHandlers/DuelChallengeDTO.ts` now has `challengerUsername?: string`.

**Step 3: Commit**

```bash
git add backend/internal/infrastructure/http/handlers/swagger_models.go \
        tma/src/api/generated/
git commit -m "feat(pvp-duel): add challengerUsername to DuelChallengeDTO swagger + regen types"
```

---

### Task 5: Add i18n keys

**Files:**
- Modify: `tma/src/i18n/locales/ru.ts:299-301`
- Modify: `tma/src/i18n/locales/en.ts:299-301`

**Step 1: Add keys to `ru.ts`**

After `challengeExpired: 'Вызов истёк',` add:

```typescript
shareAgain: 'Поделиться снова',
createNewLink: 'Создать новую ссылку',
```

**Step 2: Add keys to `en.ts`**

After `challengeExpired: 'Challenge Expired',` add:

```typescript
shareAgain: 'Share Again',
createNewLink: 'Create New Link',
```

**Step 3: Commit**

```bash
git add tma/src/i18n/locales/ru.ts tma/src/i18n/locales/en.ts
git commit -m "feat(pvp-duel): add shareAgain and createNewLink i18n keys"
```

---

### Task 6: Update `DuelLobbyView` — outgoing challenge card

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue:300-346`

**Step 1: Add `isSharing` state if not present** (already exists at line 134)

**Step 2: Replace the outgoing challenge card template section**

Find the block starting with `<!-- Outgoing Challenge -->` and ending `</div>` (currently lines ~300-346). Replace with:

```html
<!-- Outgoing Challenge (waiting for friend to accept link) -->
<div v-if="outgoingChallenges.length > 0" class="mb-4">
    <UCard
        :class="
            isOutgoingChallengeExpired
                ? 'border-gray-300 dark:border-gray-600'
                : 'border-primary-200 dark:border-primary-800'
        "
    >
        <div class="flex items-center justify-between mb-3">
            <div class="flex items-center gap-2">
                <UIcon
                    name="i-heroicons-paper-airplane"
                    :class="isOutgoingChallengeExpired ? 'text-gray-400' : 'text-primary'"
                    class="size-5"
                />
                <div>
                    <p class="font-medium text-sm">
                        {{
                            isOutgoingChallengeExpired
                                ? t('duel.challengeExpired')
                                : t('duel.waitingForFriend')
                        }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-gray-400">
                        {{ t('duel.linkExpiresIn', { time: outgoingChallengeExpiry }) }}
                    </p>
                </div>
            </div>
            <div class="flex items-center gap-2">
                <div
                    :class="
                        isOutgoingChallengeExpired
                            ? 'bg-gray-400'
                            : 'bg-primary animate-pulse'
                    "
                    class="w-2 h-2 rounded-full"
                />
                <span class="text-xs text-gray-500">
                    {{
                        isOutgoingChallengeExpired ? t('duel.expired') : t('duel.waiting')
                    }}
                </span>
            </div>
        </div>
        <!-- Action button -->
        <UButton
            icon="i-heroicons-paper-airplane"
            :color="isOutgoingChallengeExpired ? 'primary' : 'gray'"
            :variant="isOutgoingChallengeExpired ? 'solid' : 'soft'"
            size="sm"
            block
            :loading="isSharing"
            @click="handleShareToTelegram"
        >
            {{
                isOutgoingChallengeExpired
                    ? t('duel.createNewLink')
                    : t('duel.shareAgain')
            }}
        </UButton>
    </UCard>
</div>
```

**Step 3: Type check**

```bash
cd tma && pnpm run type-check
```
Expected: no errors.

**Step 4: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): add share/create button to outgoing challenge card"
```

---

### Task 7: Update `DuelLobbyView` — incoming challenge cards

**Files:**
- Modify: `tma/src/views/Duel/DuelLobbyView.vue:348-357` (pendingChallenges section)

**Step 1: Replace the pending challenges card loop**

Find the `v-for="challenge in pendingChallenges"` card block (currently has inline small buttons). Replace with block-style layout:

```html
<!-- Pending Challenges -->
<div v-if="pendingChallenges.length > 0" class="mb-4">
    <h2 class="text-sm font-semibold text-gray-600 dark:text-gray-400 mb-2">
        {{ t('duel.pendingChallenges') }}
    </h2>
    <div class="space-y-2">
        <UCard v-for="challenge in pendingChallenges" :key="challenge.id">
            <!-- Challenger identity -->
            <div class="flex items-center gap-2 mb-3">
                <UIcon name="i-heroicons-bolt" class="size-5 text-orange-500" />
                <span class="font-medium">
                    {{ challenge.challengerUsername || t('duel.challenge') }}
                </span>
            </div>
            <!-- Block buttons -->
            <div class="space-y-2">
                <UButton
                    color="green"
                    block
                    @click="() => handleAcceptChallenge(challenge.id!)"
                >
                    {{ t('duel.accept') }}
                </UButton>
                <UButton
                    color="red"
                    variant="soft"
                    block
                    @click="() => handleDeclineChallenge(challenge.id!)"
                >
                    {{ t('duel.decline') }}
                </UButton>
            </div>
        </UCard>
    </div>
</div>
```

**Step 2: Type check**

```bash
cd tma && pnpm run type-check
```
Expected: no errors.

**Step 3: Run unit tests**

```bash
cd tma && pnpm test:unit
```
Expected: all pass.

**Step 4: Commit**

```bash
git add tma/src/views/Duel/DuelLobbyView.vue
git commit -m "feat(pvp-duel): improve incoming challenge cards with challenger name and block buttons"
```

---

### Task 8: Final verification

**Step 1: Backend tests**

```bash
cd backend && go test ./...
```
Expected: all pass.

**Step 2: Frontend lint + type-check**

```bash
cd tma && pnpm lint && pnpm run type-check
```
Expected: no errors.

**Step 3: Manual smoke test (dev environment)**

Start dev environment:
```bash
# Terminal 1
cd backend && docker compose -f docker-compose.dev.yml up

# Terminal 2
cloudflared tunnel run quiz-sprint-dev

# Terminal 3
cd tma && pnpm dev
```

Open `https://dev.quiz-sprint-tma.online` → PvP Дуэль lobby:

- [ ] Outgoing challenge card: "Поделиться снова" button appears when link is active
- [ ] Outgoing expired card: "Создать новую ссылку" button appears (primary color)
- [ ] Clicking share button opens Telegram share dialog
- [ ] Incoming challenge shows challenger username (not just "Вызов")
- [ ] Accept/Decline buttons are full-width
- [ ] Accept navigates to duel play
- [ ] Decline removes the card

**Step 4: Push**

```bash
git push
```
