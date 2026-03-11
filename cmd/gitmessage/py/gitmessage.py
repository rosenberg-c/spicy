#!/usr/bin/env python3

import subprocess
import sys


def generate_commit_messages(prefix):
    """Generate a commit message using opencode based on staged git diff."""
    try:
        # Get staged diff
        print("Running: git diff --staged", file=sys.stderr)
        result = subprocess.run(
            ["git", "diff", "--staged"], capture_output=True, text=True, check=True
        )
        diff_content = result.stdout

        # Check if there are no staged changes
        if not diff_content.strip():
            print(
                "Warning: No staged changes available. Please stage your changes first.",
                file=sys.stderr,
            )
            return None

        model = "openai/gpt-5.2-codex"

        prompt = (
            "You are a senior coder: write a short commit message, one row only. "
            "Do not include the actual diff, or any other thoughts, only the commit message. "
            "Always use Capital character at the beginning of the commit message. "
            "Do not add any quotes or special characters around the response.\n\n"
            f"Diff:\n{diff_content}"
        )

        # Send the diff to opencode to generate the message
        print(f"Running: opencode run --agent build -m {model}", file=sys.stderr)
        print("Generating commit message...", file=sys.stderr)

        result = subprocess.run(
            ["opencode", "run", "--agent", "build", "-m", model, prompt],
            stdout=subprocess.PIPE,
            stderr=sys.stderr,
            text=True,
        )
        stdout = result.stdout

        if result.returncode != 0:
            raise subprocess.CalledProcessError(result.returncode, result.args)

        generated_msg = stdout.strip()

        # Conditionally prepend the prefix
        if prefix:
            return f"{prefix}: {generated_msg}"
        else:
            return generated_msg

    except subprocess.CalledProcessError as e:
        print(f"Error: {e}", file=sys.stderr)
        return None


def print_generated_message(prefix=None, copy=False):
    """Generate and print a commit message."""
    msg = generate_commit_messages(prefix)
    if msg:
        print(f"==> {msg}")
        if copy:
            subprocess.run(["pbcopy"], input=msg, text=True)


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="Generate commit messages using opencode"
    )
    parser.add_argument(
        "prefix", nargs="?", default=None, help="Prefix for the commit message"
    )
    parser.add_argument(
        "-c", "--copy", action="store_true", help="Copy result to clipboard"
    )

    args = parser.parse_args()
    print_generated_message(args.prefix, args.copy)

