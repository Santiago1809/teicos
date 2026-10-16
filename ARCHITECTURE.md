# Architecture

## Why skills are embedded at compile time

Teicos embeds its skills into the binary at build time through `go:embed`
(`internal/assets/assets.go`) instead of fetching them at runtime. This is a
deliberate security decision, not an implementation shortcut.

A skill is an instruction contract that shapes what a developer's agent writes.
If Teicos fetched skills from a remote source at runtime, updating it would mean
injecting third-party content into the developer's agents without recompilation.
That is a supply-chain surface, and it is the exact failure mode the security
context layer forbids the agent from introducing: rule 8 requires verifying a
dependency before it is added, and pinning it afterwards.

Embedding buys three properties a runtime fetch cannot:

- **Reproducibility** — a given Teicos build ships an exact, known set of skill
  bytes, so two developers on the same version get identical agent behavior.
- **Offline operation** — install and sync need no network access, so nothing is
  fetched from a host that could be compromised or unreachable.
- **Verifiability** — the skill corpus lives inside the artifact that is signed
  and released, so its integrity is covered by the release signature instead of by
  a separate trust path.

The cost is real and accepted: updating or adding a skill requires recompiling and
redistributing Teicos. That cost is paid at release time by maintainers, not at run
time inside a user's repository.

### If you are considering a runtime skill updater

Do not remove the embed to "modernize" skill distribution without addressing the
tradeoff above. A live fetcher, in the shape of `npx skills add`, moves control of
agent instructions from the released artifact to a remote host. If that is ever
required, it must arrive with its own integrity story — pinned digests, a
verifiable source, and explicit user consent — and it must never silently replace
the embedded corpus.

The design already permits this without rewriting the injector: the skills
component reads through an `fs.FS` (`internal/components/skills/inject.go`), so a
disk- or archive-backed `fs.FS` can be substituted later. Keep that seam intact.
