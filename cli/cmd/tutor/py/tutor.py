#!/usr/bin/env python3

import json
import subprocess
import sys
from enum import Enum
from typing import Optional


class Action(str, Enum):
    CONTINUE = "continue"
    EXIT = "exit"


class ValidationKey(str, Enum):
    ACTION = "action"
    REASON = "reason"
    SUGGESTIONS = "suggestions"
    SUGGESTED_FILENAME = "suggested_filename"


def spawn_agent_response(
    model: str, prompt: str, verbose: bool = False
) -> tuple[Optional[str], Optional[Exception]]:
    result = subprocess.run(
        ["opencode", "run", "--agent", "build", "-m", model, prompt],
        stdout=subprocess.PIPE,
        stderr=sys.stderr if verbose else subprocess.DEVNULL,
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


def build_validation_prompt(user_input: str) -> str:
    return f"""You are a senior technical writer and educator. Analyze the following user request for a tutorial:

"{user_input}"

Determine if this request is specific and clear enough to create a useful tutorial, or if it's too ambiguous and needs clarification.

Respond ONLY with a valid JSON object (no markdown, no extra text) in this exact format:

If the request is specific enough:
{{
  "{ValidationKey.ACTION.value}": "{Action.CONTINUE.value}",
  "{ValidationKey.REASON.value}": "brief explanation of your decision",
  "{ValidationKey.SUGGESTED_FILENAME.value}": "short-descriptive-name.md"
}}

If the request is too ambiguous:
{{
  "{ValidationKey.ACTION.value}": "{Action.EXIT.value}",
  "{ValidationKey.REASON.value}": "brief explanation of your decision",
  "{ValidationKey.SUGGESTIONS.value}": ["clarifying question 1", "clarifying question 2"]
}}

Guidelines for suggested_filename:
- Use lowercase with hyphens (kebab-case)
- Keep it short but descriptive (2-5 words max)
- Must end with .md
- Examples: "ffmpeg-video-conversion.md", "pandas-csv-guide.md", "echo-command-basics.md"

Examples of decisions:
- "how does ffmpeg work" -> too broad, suggest: "how to convert video formats", "how to extract audio", etc.
- "how to convert mp4 to webm using ffmpeg" -> specific enough, continue, suggest: "ffmpeg-mp4-to-webm.md"
- "explain python" -> too vague, suggest: "which aspect of Python", "what's your experience level", etc.
- "how to read a CSV file in Python using pandas" -> specific enough, continue, suggest: "pandas-csv-reading.md"
- "docker" -> too vague, suggest: "docker basics", "docker compose", "dockerfile best practices", etc.
- "echo command" -> specific enough, continue, suggest: "echo-command-guide.md"

Think carefully about whether the request has enough context and specificity to create a useful, focused tutorial."""


def validate_user_input(
    model: str, user_input: str, verbose: bool = False
) -> tuple[Optional[dict], Optional[Exception]]:
    validation_prompt = build_validation_prompt(user_input)
    response, err = spawn_agent_response(
        model=model, prompt=validation_prompt, verbose=verbose
    )

    if err:
        return None, err

    assert response is not None

    try:
        # Parse the JSON response
        validation_result = json.loads(response)

        # Validate the structure
        if (
            ValidationKey.ACTION.value not in validation_result
            or ValidationKey.REASON.value not in validation_result
        ):
            return None, ValueError(
                "Invalid validation response structure: missing required fields"
            )

        if validation_result[ValidationKey.ACTION.value] not in [
            Action.CONTINUE.value,
            Action.EXIT.value,
        ]:
            return None, ValueError(
                f"Invalid action value: {validation_result[ValidationKey.ACTION.value]}"
            )

        # Validate action-specific fields
        if validation_result[ValidationKey.ACTION.value] == Action.CONTINUE.value:
            if ValidationKey.SUGGESTED_FILENAME.value not in validation_result:
                return None, ValueError(
                    "Missing required field 'suggested_filename' for continue action"
                )

        return validation_result, None

    except json.JSONDecodeError as e:
        return None, ValueError(
            f"Failed to parse JSON response: {e}\nResponse was: {response}"
        )


def build_tutorial_prompt(user_input: str) -> str:
    return (
        "You are a senior coder. "
        "Write a tutorial to answer the user question, as detailed as you can. "
        "The response must be valid markdown. "
        f"User input:\n{user_input}"
    )


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="Generate tutorial", usage="%(prog)s [question...]"
    )
    parser.add_argument(
        "question",
        nargs="*",
        help="The question for the tutorial (e.g., 'how does echo command work')",
    )
    parser.add_argument(
        "-v",
        "--verbose",
        action="store_true",
        help="Show verbose output from agent commands",
    )
    args = parser.parse_args()

    print("== Genretae Tutorial ==")

    model = "openai/gpt-5.2"

    # Get user input from CLI args or prompt
    if args.question:
        user_input = " ".join(args.question)
        print(f"Question: {user_input}")
    else:
        user_input, err = get_user_input()
        if err:
            print(f"Error getting user input: {err}", file=sys.stderr)
            sys.exit(1)
        assert user_input is not None

    # Validate user input first
    print("Validating input...", file=sys.stderr)
    validation_result, err = validate_user_input(
        model=model, user_input=user_input, verbose=args.verbose
    )
    if err:
        print(f"Error validating user input: {err}", file=sys.stderr)
        sys.exit(1)
    assert validation_result is not None

    # Check if we should exit based on validation
    if validation_result[ValidationKey.ACTION.value] == Action.EXIT.value:
        print(f"\n{validation_result[ValidationKey.REASON.value]}", file=sys.stderr)
        if (
            ValidationKey.SUGGESTIONS.value in validation_result
            and validation_result[ValidationKey.SUGGESTIONS.value]
        ):
            print("\nSuggestions:", file=sys.stderr)
            for suggestion in validation_result[ValidationKey.SUGGESTIONS.value]:
                print(f"  - {suggestion}", file=sys.stderr)
        sys.exit(0)

    # Ask for save location after validation passes
    suggested_filename = validation_result.get(
        ValidationKey.SUGGESTED_FILENAME.value, "tutorial.md"
    )
    output_path = (
        input(f"Save to file (default: {suggested_filename}) => ").strip()
        or suggested_filename
    )

    # Continue with tutorial generation
    print("Generating tutorial...", file=sys.stderr)
    prompt = build_tutorial_prompt(user_input)
    tutorial, err = spawn_agent_response(model=model, prompt=prompt, verbose=args.verbose)
    if err:
        print(f"Error spawning agent response: {err}", file=sys.stderr)
        sys.exit(1)
    assert tutorial is not None

    fpath, err = save_to_disk(content=tutorial, output_path=output_path)
    if err:
        print(f"Error saving to disk: {err}", file=sys.stderr)
        sys.exit(1)

    print(f"Saved to: {fpath}")
