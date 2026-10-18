# Security policy

## Reporting a vulnerability

Please report security issues privately through
[GitHub security advisories](https://github.com/mathieuHa/resticgui/security/advisories/new),
not in public issues. Include what an attacker needs (network access, a
login, a crafted repository...), the steps to reproduce, and the version
(`resticgui version`).

You will get an answer within a week. Fixes are released as soon as they
are ready, and the advisory is published with the release that fixes it,
crediting you unless you prefer otherwise.

## Supported versions

Only the latest release gets security fixes. The container image is
rebuilt every week with the latest Alpine and restic fixes.

## What ResticGUI protects

ResticGUI reads every file it backs up and writes wherever it restores, so
access to it is close to access to the machine. The
[security page](https://docs.resticgui.com/reference/security) explains how
to run it safely, and the
[security review](https://docs.resticgui.com/reference/security-review)
lists what was checked, what was found and how it was fixed.
