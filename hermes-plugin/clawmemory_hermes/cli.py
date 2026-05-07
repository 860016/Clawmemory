from __future__ import annotations

import argparse
import json
import sys

from . import ClawMemoryProvider, create_provider


def main() -> None:
    parser = argparse.ArgumentParser(
        prog="clawmemory-hermes",
        description="ClawMemory Memory Provider for Hermes Agent",
    )
    sub = parser.add_subparsers(dest="command")

    save_p = sub.add_parser("save", help="Save a memory")
    save_p.add_argument("key", help="Memory key")
    save_p.add_argument("value", help="Memory value")
    save_p.add_argument("--layer", default="episodic")
    save_p.add_argument("--type", default="knowledge", dest="memory_type")

    recall_p = sub.add_parser("recall", help="Search memories")
    recall_p.add_argument("query", help="Search query")
    recall_p.add_argument("--limit", type=int, default=5)

    context_p = sub.add_parser("context", help="Get context for a query")
    context_p.add_argument("query", help="Context query")
    context_p.add_argument("--limit", type=int, default=5)

    reason_p = sub.add_parser("reason", help="Run dialectic reasoning")
    reason_p.add_argument("query", help="What to reason about")
    reason_p.add_argument("--depth", type=int, default=1)
    reason_p.add_argument("--level", default="low")

    test_p = sub.add_parser("test", help="Test connection to ClawMemory backend")

    parser.add_argument("--base-url", default="http://localhost:8765")
    parser.add_argument("--api-key", default="")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        sys.exit(1)

    config = {
        "base_url": args.base_url,
        "api_key": args.api_key,
    }
    provider = create_provider(config)

    try:
        if args.command == "save":
            provider.save(args.key, args.value, layer=args.layer, memory_type=args.memory_type)
            print(f"Saved: {args.key}")

        elif args.command == "recall":
            results = provider.recall(args.query, args.limit)
            for r in results:
                print(f"  [{r.get('source', '?')}] {r.get('key', '?')}: {r.get('value', '')[:100]}")

        elif args.command == "context":
            ctx = provider.get_context(args.query, args.limit)
            print(ctx or "No context found.")

        elif args.command == "reason":
            result = provider.session.reason(args.query, args.depth, args.level)
            print(result or "No reasoning result.")

        elif args.command == "test":
            results = provider.recall("test", 1)
            print(f"Connection OK. Found {len(results)} test results.")

    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    finally:
        provider.close()


if __name__ == "__main__":
    main()
