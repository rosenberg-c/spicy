#!/usr/bin/env python3

import subprocess
import sys


def spawn_agent_response(model, prompt):
    try:
        result = subprocess.run(
            ["opencode", "run", "--agent", "build", "-m", model, prompt],
            stdout=subprocess.PIPE,
            stderr=sys.stderr,
            text=True,
        )
        stdout = result.stdout

        if result.returncode != 0:
            raise subprocess.CalledProcessError(result.returncode, result.args)

        return stdout.strip()

    except subprocess.CalledProcessError as e:
        print(f"Error: {e}", file=sys.stderr)
        return None


def save_to_disk(content, output_path):
    try:
        from pathlib import Path
        import os
        import tempfile

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
        return str(path)

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return None


def get_user_input():
    user_input = input("-- input: ").strip()
    if not user_input:
        raise ValueError("You must provide an input")
    return user_input


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(description="Generate tutorial")
    args = parser.parse_args()

    print("== Genretae Tutorial ==")
    output_path = (
        input("Save to file (default: tutorial.md) => ").strip() or "tutorial.md"
    )

    model = "openai/gpt-5.2"
    user_input = get_user_input()
    prompt = (
        "You are a senior coder:"
        "Write a tutorial to answer the user question, as detailed as you can. "
        "The response must be valid markdown. "
        f"User input:\n{user_input}"
    )

    print("Generating tutorial...", file=sys.stderr)
    tutorial = spawn_agent_response(model=model, prompt=prompt)

    fpath = save_to_disk(content=tutorial, output_path=output_path)
    print(f"Saved to: {fpath}")
