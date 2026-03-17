# Tutorial: What “rc” Means in `.bashrc` and Other Config Files

## Short Answer
**`rc` stands for “run commands.”** In Unix-like systems, `rc` files are scripts or configuration files that are executed (or “run”) when a program starts.

---

## Where “rc” Comes From
The term dates back to early Unix and even older systems (like CTSS). Over time, “rc” became the conventional suffix for *startup* or *initialization* files.  

So when you see:

- `.bashrc`
- `.vimrc`
- `.gitrc` (or Git’s `~/.gitconfig`)
- `.screenrc`

…they’re all **startup or runtime configuration files** that the program reads when it launches.

---

## How `.bashrc` Works (Concrete Example)
Bash reads different config files depending on how you start it:

- **Interactive, non‑login shells** → reads `~/.bashrc`
- **Login shells** → reads `~/.bash_profile` or `~/.profile`

So `.bashrc` is the **“run commands for Bash”** file for interactive sessions.

A typical `.bashrc` might contain:

```bash
# ~/.bashrc
export EDITOR=vim
alias ll='ls -lah'
PS1='\u@\h:\w\$ '
```

When a new terminal starts, Bash runs those commands automatically.

---

## The Pattern Across Other Tools
Many programs follow the same “rc” convention:

| File | Program | Purpose |
|------|---------|---------|
| `~/.vimrc` | Vim | Editor configuration |
| `~/.zshrc` | Zsh | Shell configuration |
| `~/.tmux.conf` (not rc) | Tmux | Startup config (same idea) |
| `~/.screenrc` | GNU Screen | Startup config |
| `~/.inputrc` | Readline | Input editing rules |

Even when the suffix isn’t literally `rc`, the concept is the same: a file that is **read at startup to configure behavior**.

---

## Why It’s a Hidden File
On Unix-like systems, files beginning with `.` are hidden. This helps keep your home directory clean, since configuration files are not meant to be edited constantly.

You can list them with:

```bash
ls -a
```

---

## Quick Mental Model
When you see **`rc`**, think:

- **R**un **C**ommands  
- **Startup / initialization / configuration**  
- **“This file is read when the program starts”**

---

## Next Steps (Optional)
If you want to explore further:

1. Open your `~/.bashrc` and see what’s inside:
   ```bash
   nano ~/.bashrc
   ```
2. Try adding a custom alias, then reload:
   ```bash
   source ~/.bashrc
   ```
3. Compare with other rc files like `~/.vimrc` or `~/.zshrc`.

If you want help customizing your shell config, just say the word.
