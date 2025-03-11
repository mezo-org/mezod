# Mezo Contribution Guide

We appreciate your interest in contributing to the source code. Contributions
from anyone are always welcome, and even the smallest improvements are
greatly valued.

If you would like to contribute to Mezo, start by forking the repository, making
your changes, committing them, and submitting a pull request for the maintainers
to review and merge into the main codebase. For more significant modifications,
please reach out to the developers on the [Mezo Discord Server](https://discord.mezo.org/)
beforehand. This helps ensure your changes align with the project's overall
vision and allows you to receive early feedback, making both your work and our
review process smoother and more efficient.

## Developer Documentation

The Mezo client can be run locally for development purposes. The Developer
Documentation is available in the [`docs` directory](https://github.com/mezo-org/mezod/tree/main/docs).

## Developer Tooling

### Continuous Integration

Mezo uses [Github Actions](https://github.com/mezo-org/mezod/actions) for
continuous integration. All jobs must be green to merge a PR.

### Pre-commit

Pre-commit is a tool to install hooks that check code before commits are made.
It can be helpful to install this, to automatically run linter checks and avoid
pushing code that will not be accepted. Follow the
[installation instructions here](https://pre-commit.com/), and then run
`pre-commit install` to install the hooks.

### Linting

Linters and formatters for Solidity, JavaScript, and Go code are set up and run
automatically as part of pre-commit hooks. These are checked again in CI builds
to ensure they have been run and are passing.

### Commit signing

All commits [must be signed](https://help.github.com/en/articles/about-commit-signature-verification).

## Commit Messages

When composing commit messages, please follow the general guidelines listed in
[Chris Beams’s How to Write a Git Commit Message](https://cbea.ms/git-commit/).
Many editors have git modes that will highlight overly long first lines of
commit messages, etc. The GitHub UI itself will warn you if your commit summary
is too long, and will auto-wrap commit messages made through the UI to 72
characters.

The above goes into good commit style. Some additional guidelines do apply,
however:

* The target audience of your commit messages is always "some person 10 years
  from now who never got a chance to talk to present you" (that person could be
  future you!).
* Commit messages with a summary and no description should be very rare. This
  means you should probably break any habit of using `git commit -m`.
* A fundamental principle that informs our use of GitHub: assume GitHub will
  someday go away, and ensure git has captured all important information about
  the development of the code. Commit messages are the piece of knowledge that
  is second most likely to survive tool transitions (the first is the code
  itself); as such, they must stand alone. Do not reference tickets or issues
  in your commit messages. Summarize any conclusions from the issue or ticket
  that inform the commit itself, and capture any additional reasoning or context
  in the merge commit.
* Make your commits as atomic as you can manage. This means each commit contains
  a single logical unit of work.
* Run a quick `git log --graph --all --oneline --decorate` before pushing.
  It’s much easier to fix typos and minor mistakes locally.
