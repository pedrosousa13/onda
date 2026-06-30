# radio — UX Pass 2 Plan

Consolidates the feedback gathered while testing the v1 TUI. Grouped into phases;
each phase ends in a buildable, testable, committable state. Branch: `feat/v1-mvp`.

## Already shipped this session (for context)
- Mirror racing (init speed), duplicate-stream merge (name+country, strip
  parens/brackets/codec tokens), HiFi/lossless ranking + labels, bitrate chooser,
  per-quality dedupe, `UNKNOWN`→`—`, themes, Now-Playing hero, Popular default.

---

## Phase 1 — Now-Playing panel + metadata polish (quick wins)

**1a. Chooser clarity.** Move the volume meter to the now-playing title line
(right-aligned); put the bitrate chooser on its own line, left-aligned, with the
**active quality bracketed in the accent color** (`[192k]`) so it's unmistakable
even at a glance / in screenshots. Fixes "can't tell what's selected, too far apart."

**1b. Sanitize song metadata.** Some streams push XML as ICY metadata (Radio
Comercial / Dalet: `<RadioInfo>…<DB_DALET_TITLE_NAME>…`). Detect markup and:
extract artist/title from known tags → show "Artist — Title"; otherwise suppress
and show "live". Never render raw XML.
- Verify: unit test `sanitizeTitle` for the Dalet sample, a clean ICY title, and an unknown-XML fallback.

**1c. Navigation discoverability.** Add `esc home` to the footer; `esc` from any
view returns to Home (Phase 4) / Popular (until then). Answers "how to go back?".

---

## Phase 2 — Search by name + country + tags, with fuzzy ranking

**2a. Multi-field search.** Today search only queries Radio Browser by `name`.
Issue parallel RB calls — `byname`, `bytag`, `bycountry` (raced/merged via the
existing mirror logic), dedupe through `GroupRecords`. One box, searches all three.

**2b. Fuzzy ranking.** Rank merged results by a fuzzy score against the query
(typo/word-order tolerant), so the best matches sort to the top. Lightweight
scorer (or `sahilm/fuzzy`); applied client-side after merge.
- Verify: unit test that "jazz" surfaces FIP Jazz; that a country and a tag query each return hits; that fuzzy ordering puts close matches first.

---

## Phase 3 — Smarter de-duplication (best-effort, with honest limits)

Real cases seen: `Antena 3` vs `RTP Antena 3` vs `Antena 3 - Main` vs
`Antena 3 Madeira` (region) vs `Radio Antena 3` (Ecuador, genuinely different).

**3a. Normalize more:** strip a trailing `- Main`/`Main`/`HD`; strip a small,
curated set of **broadcaster prefixes** (e.g. `RTP `, `BBC `, `NPR `, `RAI `) when
matching, so `RTP Antena 3` ≈ `Antena 3`. Keep region words (`Madeira`) distinct.

**3b. Honest limit (documented).** Radio Browser has **no canonical station IDs**;
it's crowd-sourced, so perfect grouping is impossible. We do best-effort heuristics
and accept a few stragglers rather than risk merging genuinely different stations.
A `log`/README note will state this.
- Verify: tests for `RTP Antena 3`≈`Antena 3` merge; `…Madeira` and Ecuador stay separate.

---

## Phase 4 — Home view (the headline UX change)

A simplified landing screen, replacing "land straight in the Popular list":

```
 radio · wander the world                                              home

 ╭─ now playing ──────────────────────────────────────────────────────────╮
 │ ♫ FIP                                              ▮▮▮▮▮▮▮▮▯▯ 80%        │
 │   Khruangbin — Maria También                                            │
 │   quality  [HiFi] 192k 128k                                             │
 ╰──────────────────────────────────────────────────────────────────────────╯

 favorites
 ▌ KEXP                                                  United States · 128k
   BBC World Service                                      United Kingdom · 96k
   (no favorites yet — press / to search, ★ to save one)

 /  search     P  popular     F  favorites     a  add     ,  settings    q  quit
```

- **Home = now-playing hero (full) + favorites list** (falls back to a short
  Popular preview when you have no favorites yet). This is the "bare minimum" view.
- From Home: `enter` plays the highlighted favorite; `/` search, `P` popular (full
  list), `F` favorites (full), `a` add, `,` settings.
- **`esc` returns to Home** from anywhere. Home is the app's center of gravity.
- The rich Popular/Search/Favorites lists remain one key away (current hero list).
- Verify: model tests for view transitions (home→search→esc→home; play-from-home);
  render-gallery frame for Home.

---

## Suggested order & sizing
1. **Phase 1** (small, high polish payoff) — do first.
2. **Phase 4 Home** (medium, biggest UX lift) — do next; it reframes navigation.
3. **Phase 2 search** (medium) — name+country+tag+fuzzy.
4. **Phase 3 dedup** (small-medium, diminishing returns) — last.

Each phase: TDD where logic exists (sanitize, search merge, dedup, transitions),
render-gallery check for visuals, commit, push (updates PR #1).
