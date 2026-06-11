# Syncraft

> _Collaborative Minecraft worlds without a shared server — sync your saves like code._

---

## What is this?

Syncraft is a lightweight tool that lets two or more players share a Minecraft world asynchronously. Instead of transferring entire world files, it generates and syncs only what changed — keeping bandwidth low and the workflow simple.

## Problem it solves

Co-op Minecraft typically requires a live shared server, which most friend groups don't want to maintain. The alternative — manually swapping world folders via a shared drive or Git — works, but transfers the entire world every time, far more data than is actually necessary. Syncraft solves this by diffing at the block level, so only real changes move over the wire.

## How it works

1. Play a session of Minecraft locally.
2. Save your world.
3. Syncraft generates a diff of what changed since the last sync.
4. Push the diff to the Syncraft backend.
5. Other players pull the latest diff before their next session and apply it locally.

Diffs are computed at the block level (not chunk level), so the sync surface is as small as possible. A backend server handles storage and serving of diff files.

## Tech stack (tentative)

<!--
How will diffs be generated? (binary diffing of .mca region files, or a custom block-level format?)
What language / framework for the CLI or app?
What does the backend look like? (e.g. small REST API + object storage)
-->

## Open questions

- **Conflict handling** — if two players edit the same block in separate sessions, how is that resolved? Per-block diffing makes this granular, but a merge strategy still needs to be defined.
- **Minecraft integration** — running standalone is simpler to build; a launcher or mod integration would be smoother for users but adds complexity. TBD.
- **Backend storage model** — does the server store full snapshots, a chain of diffs, or both? Affects how far back you can roll back and how expensive a fresh sync is for a new player.

# Minecraft Files to be Checked:

- level.dat — a single NBT file containing world-level metadata like seed, time, gamerules, spawn point
- playerdata/\*.dat — one NBT file per player, storing inventory, health, position etc
- entities/\*.mca — same region file format as chunks but stores entity data separately (added in 1.17)
- poi/\*.mca
