# Contributing

The code is public and MIT. Development is done by the maintainers.

We do not accept unsolicited pull requests.

A PR is a review obligation. In 2026 that queue is mostly drive-by and AI-generated patches that look finished and are not. Same move as [tldraw](https://github.com/tldraw/tldraw/issues/7695), [fzf](https://github.com/junegunn/fzf/issues/4752), [Documenso](https://github.com/documenso/documenso/issues/3026), [Baserow](https://github.com/baserow/baserow/issues/5598), and [Ladybird](https://ladybird.org/posts/changing-how-we-develop-ladybird/): keep the source open, stop staffing a public merge queue.

Fork it. Run it. Audit it. Open an issue. If the idea fits, we will write the change.

## What to open

GitHub issues are the contribution path.

- Bug reports (what you ran, what you expected, what you got)
- Ideas and feature requests
- Design discussion
- Security reports
- A sample workspace, a screenshot, a failing command

A good issue reads like a spec. We will close it, ask questions, or build against it.

You can paste a snippet or link a fork to show a point. That is illustration, not a patch queue. We will not review issue-attached diffs as if they were PRs.

## What we will close

- Unsolicited pull requests. We will close them and point here. Open an issue instead.
- Skill files (`AGENTS.md`, `SKILL.md`, `recs agent install`)
- Gmail / IMAP / Outlook, cloud sync, login, an in-app model or embeddings, Cosmopolitan / APE

Those last two lists are permanent non-goals, not later work.

## If you already opened a PR

It is not a judgment on the idea. File the same thing as an issue. If we want the change, we will implement it.

## Building it yourself

```
go test ./...
go build -o recs ./cmd/recs
```

Start from `recs --help`. Help is the prompt. `--json` is the contract.
