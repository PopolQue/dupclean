# CLI Interactive UI

When scanning for duplicates in CLI mode, DupClean provides an interactive interface to review and manage duplicate groups.

## Interaction

For each duplicate group, you will see all copies with their:

- Filename
- Full path
- Size
- Last modified date

Then choose which copy to **keep** (others go to Trash), or **skip** the group.

``` CLI
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Group 1 of 4  (identical audio content)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  [1]  kick_drum_01.wav
       /Users/you/Samples/drums/kick_drum_01.wav
       Size: 1.2 MB  Modified: 2024-03-15 11:22

  [2]  Kick Hard v2 FINAL.wav
       /Users/you/Desktop/old stuff/Kick Hard v2 FINAL.wav
       Size: 1.2 MB  Modified: 2023-09-02 09:14

  Keep which file? (1-2)  [s]kip  [a]ll skip  [q]uit
  > 1
  ✓ Keeping: kick_drum_01.wav
  ⛌ Deleted: Kick Hard v2 FINAL.wav
```

## Controls

| Input | Action |
| ----- | ------ |
| `1`, `2`, ... | Keep that file, trash the rest |
| `s` or Enter | Skip this group |
| `a` | Skip all remaining groups |
| `q` | Quit |
