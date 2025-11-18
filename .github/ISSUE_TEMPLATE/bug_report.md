---
name: Bug Report
about: Report a bug or unexpected behavior
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Go to '...'
2. Execute '...'
3. Observe '...'

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

A clear and concise description of what actually happened.

## Code Sample

```go
// Minimal code sample that reproduces the issue
package main

import (
    "github.com/moffa90/go-cyacd/bootloader"
    "github.com/moffa90/go-cyacd/cyacd"
)

func main() {
    // Your code here
}
```

## Environment

- **go-cyacd version:** [e.g., v1.0.0]
- **Go version:** [e.g., 1.23.0]
- **OS:** [e.g., Ubuntu 22.04, Windows 11, macOS 14]
- **Architecture:** [e.g., amd64, arm64]
- **Device/Hardware:** [e.g., PSoC 4, PSoC 5LP]
- **Communication method:** [e.g., USB HID, UART]

## Error Output

```
Paste any error messages, stack traces, or logs here
```

## Additional Context

Add any other context about the problem here:

- Does this happen consistently or intermittently?
- Did this work in a previous version?
- Any workarounds you've found?
- Related issues or discussions?

## Firmware File

If applicable:
- [ ] I can share the .cyacd firmware file (attach or link)
- [ ] The issue is reproducible with any firmware file
- [ ] I cannot share the firmware file due to confidentiality

## Checklist

- [ ] I have searched existing issues to avoid duplicates
- [ ] I have tested with the latest version
- [ ] I have provided a minimal code sample
- [ ] I have included the full error output
- [ ] I have specified my environment details
