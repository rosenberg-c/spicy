#!/usr/bin/env python3

import subprocess
import sys
from typing import Optional


def spawn_agent_response(
    model: str, prompt: str
) -> tuple[Optional[str], Optional[Exception]]:
    result = subprocess.run(
        ["opencode", "run", "--agent", "build", "-m", model, prompt],
        stdout=subprocess.PIPE,
        stderr=sys.stderr,
        text=True,
    )
    stdout = result.stdout

    if result.returncode != 0:
        error = subprocess.CalledProcessError(result.returncode, result.args)
        return None, error

    return stdout.strip(), None


def save_to_disk(
    content: str, output_path: str
) -> tuple[Optional[str], Optional[Exception]]:
    from pathlib import Path
    import os
    import tempfile

    try:
        path = Path(os.path.expanduser(output_path)).resolve()
        if str(path.parent):
            path.parent.mkdir(parents=True, exist_ok=True)

        if not content.endswith("\n"):
            content += "\n"

        with tempfile.NamedTemporaryFile(
            mode="w",
            encoding="utf-8",
            newline="\n",
            delete=False,
            dir=str(path.parent),
            prefix=f".{path.name}.",
            suffix=".tmp",
        ) as tmp:
            tmp.write(content)
            tmp_path = Path(tmp.name)

        tmp_path.replace(path)
        return str(path), None

    except Exception as e:
        return None, e


def get_user_input() -> tuple[Optional[str], Optional[ValueError]]:
    user_input = input("-- input: ").strip()
    if not user_input:
        return None, ValueError("You must provide an input")
    return user_input, None


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Generate tutorial")
    args = parser.parse_args()

    print("== Genretae Tutorial ==")
    output_path = (
        input("Save to file (default: tutorial.md) => ").strip() or "tutorial.md"
    )

    model = "openai/gpt-5.2"
    user_input, err = get_user_input()
    if err:
        print(f"Error getting user input: {err}", file=sys.stderr)
        sys.exit(1)

    prompt = (
        "You are a senior coder:"
        "Write a tutorial to answer the user question, as detailed as you can. "
        "The response must be valid markdown. "
        f"User input:\n{user_input}"
    )

    print("Generating tutorial...", file=sys.stderr)
    tutorial, err = spawn_agent_response(model=model, prompt=prompt)
    if err:
        print(f"Error spawning agent response: {err}", file=sys.stderr)
        sys.exit(1)
    assert tutorial is not None

    fpath, err = save_to_disk(content=tutorial, output_path=output_path)
    if err:
        print(f"Error saving to disk: {err}", file=sys.stderr)
        sys.exit(1)

    print(f"Saved to: {fpath}")
