from __future__ import annotations

import sys
import logging
import argparse
import json
import time
import random
import socket


from dataclasses import dataclass, asdict
from datetime import datetime, timezone
from typing import Literal

Level = Literal ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]

@dataclass(frozen=True)
class Runtime:
    app_name: str = "loggenerator"
    dev_mode: bool = False
    wait: float = 0.25
    logs: int = 10000
    host: str = "127.0.0.1"
    port: int = 9000

@dataclass(frozen=True)
class LogEvent:
    timestamp: str
    level: Level
    app_name: str
    logger_name: str
    function_name: str
    message: str

def setup_logging() -> None:
    root_logger = logging.getLogger()
    root_logger.setLevel(logging.DEBUG)

    root_logger.handlers.clear()
    
    formatter = logging.Formatter(
            fmt="%(asctime)s %(levelname)s %(name)s.%(funcName)s() %(message)s",
            datefmt="%Y-%m-%d %H:%M:%S",
    )
    stderr_handler = logging.StreamHandler(sys.stderr)
    stderr_handler.setLevel(logging.DEBUG)
    stderr_handler.setFormatter(formatter)
    root_logger.addHandler(stderr_handler)
    return None


def generate(runtime: Runtime) -> LogEvent:
    levels: list[Level] = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
    messages = [
        "User login succeeded",
        "Database connection slow",
        "Disk usage high",
        "Unexpected input received",
        "Unhandled exception in worked",
    ]

    return LogEvent(
        timestamp=datetime.now(timezone.utc).isoformat(),
        level=random.choice(levels),
        app_name=runtime.app_name,
        logger_name=__name__,
        function_name="generate",
        message=random.choice(messages),
    )

def loop(runtime: Runtime) -> int:
    logger = logging.getLogger(__name__)

    logger.info("Starting to generate")
    with socket.create_connection((runtime.host, runtime.port)) as sock:
        logger.info("Connection setup ..")
        for _ in range(runtime.logs):
            event = generate(runtime)
            payload = json.dumps(asdict(event)) + "\n"
            sock.sendall(payload.encode("utf-8"))

            logger.info("Sent event: %s", event.message)
            if runtime.dev_mode:
                time.sleep(runtime.wait)
    logger.info("Exiting ..")
    return 0

def main() -> int:
    """
    Generates logs and sends them to ports
    """
    app_name = "loggenerator"
    parser = argparse.ArgumentParser(prog = app_name)

    parser.add_argument("--dev", action="store_true", help="Enable dev conveniences")
    parser.add_argument("--logs", type=int, default=10000, help="Enable dev conveniences")
    parser.add_argument("--host", type=str, default="127.0.0.1", help="Target host")
    parser.add_argument("--port", type=int, default=9000, help="Target port")

    args = parser.parse_args(sys.argv[1:])
    try:
        runtime = Runtime(
            app_name = app_name,
            dev_mode = args.dev,
            logs = int(args.logs),
            host = args.host,
            port = int(args.port),
        )
        setup_logging()
        return loop(runtime)

    except KeyboardInterrupt:
        return 130
    except Exception:
        logging.getLogger(__name__).exception("Fatal error")
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
