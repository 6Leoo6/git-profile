# git-profile

A small CLI tool for switching Git author identities. Useful when you work across multiple GitHub accounts — for example, a work and a personal one — and want to make sure every commit is attributed to the right one.

```
  Profiles
    personal          Your Name <you@personal.com>
    work              Your Name <you@work.com>

  Active identity
    local             Your Name <you@work.com>
    global            Your Name <you@personal.com>
```

Profiles are stored in `~/.git-profiles.json`. Switching applies to the **current repository by default**, leaving your global identity untouched unless you explicitly pass `--global`.

---

## Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- Git on your `PATH`

### Build from source

```sh
git clone https://github.com/yourusername/git-profile.git
cd git-profile
go mod tidy
go build -o git-profile .        # macOS / Linux
go build -o git-profile.exe .    # Windows
```

Then move the binary somewhere on your `PATH`:

```sh
# macOS / Linux
mv git-profile /usr/local/bin/

# Windows (PowerShell) — assuming $HOME\scripts is already on PATH
Move-Item git-profile.exe "$HOME\scripts\git-profile.exe"
```

### Verify

```sh
git-profile --version
```

---

## Setup

On first run, `git-profile` creates a sample config at `~/.git-profiles.json` and exits:

```json
{
  "profiles": {
    "personal": {
      "name": "Your Name",
      "email": "you@personal.com"
    },
    "work": {
      "name": "Your Name",
      "email": "you@work.com"
    }
  }
}
```

Edit it with your real details, then run `git-profile` again. You can also manage profiles directly from the CLI — see [Managing profiles](#managing-profiles) below.

> **Note:** The email must match a verified address on the corresponding GitHub account for commits to be correctly attributed. You can check this under **GitHub → Settings → Emails**.

---

## Usage

### Show status

```sh
git-profile
```

Lists all saved profiles and your currently active identity. If you're inside a repository, it shows the local override and the global fallback separately. Warns you if the active identity doesn't match any saved profile.

### Switch profile

```sh
git-profile work
```

Applies the `work` profile to the **current repository** by writing to `.git/config`. Must be run from inside a Git repository.

```sh
git-profile work --global
```

Applies the profile to `~/.gitconfig`, affecting all repositories that don't have a local override.

### Managing profiles

```sh
# Add or update a profile
git-profile add freelance "Your Name" you@freelance.com

# Remove a profile
git-profile remove freelance
```

You can also edit `~/.git-profiles.json` directly in any text editor.

### Other commands

```sh
git-profile --help       # show usage
git-profile --version    # show version
```

---

## How Git identity resolution works

Git resolves `user.name` and `user.email` in this order:

1. **Local** — `.git/config` inside the repository
2. **Global** — `~/.gitconfig`
3. **System** — `/etc/gitconfig`

`git-profile` writes locally by default so your global config stays as a clean fallback. Running `git-profile` with no arguments always shows you exactly which level is active, so there's no guessing.

---

## Fixing past commits

If you already pushed commits with the wrong identity, you can rewrite them.

**Last N commits:**

```sh
git rebase -i HEAD~N
# mark each commit as 'edit', then for each:
git commit --amend --author="Your Name <you@work.com>" --no-edit
git rebase --continue
```

**Many commits at once** — requires [`git-filter-repo`](https://github.com/newren/git-filter-repo):

```sh
git filter-repo --email-callback '
    return b"you@work.com" if email == b"you@personal.com" else email
'
```

> ⚠️ Rewriting history changes commit SHAs. You will need to force-push (`git push --force`) and coordinate with any collaborators.

---

## License

MIT