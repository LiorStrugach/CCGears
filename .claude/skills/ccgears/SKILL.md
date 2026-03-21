---
name: ccgears
description: >
  Switch to the CCGears preset manager. Automatically exits this session
  and opens CCGears. Session resumes after preset switch.
allowed-tools: Bash(touch *), Bash(export *), Bash(ccgears *), Bash(kill *)
---

# CCGears — Switch Presets

When this skill is invoked, IMMEDIATELY run this single Bash command.
Do NOT ask questions. Do NOT present options. Just execute this command:

```bash
touch /tmp/.ccgears-switch && export PATH="$PATH:$HOME/go/bin" && ccgears list && kill -INT $PPID
```

This will list the available presets, then exit this session. CCGears will
open automatically and this session will resume after the preset switch.
