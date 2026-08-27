#!/usr/bin/env python3
from pathlib import Path
import re

root = Path(__file__).resolve().parents[1]
text_files = [
    path
    for path in root.rglob("*")
    if path.is_file()
    and path.relative_to(root).parts[0] not in {"preflight", ".git"}
    and path.name != "PACKAGE_SHA256SUMS"
    and "tests" not in path.relative_to(root).parts
]
joined = "\n".join(path.read_text(errors="strict") for path in text_files)

assert not re.search(r"\bOn Railway\b", joined, re.I)
assert "OWNER/REPOSITORY" not in joined
assert (
    joined.count(
        "railway domain --service influxdb3-core --port 8181 --json"
    )
    == 1
)
for required in [
    "InfluxDB 3 Core",
    "3.10.0",
    "sha256:b3e577f38c19963597170d8850a3a7f77af8f0cfa866c64cd13e5de0f238e114",
    "${{secret(64)}}",
    "PORT=8181",
    "/healthz",
    "/data",
    "1500:1500",
    "Serverless",
    "one-way",
    "offline",
    "not endorsed",
]:
    assert required in joined, required

for forbidden in [
    r"(?i)\bproduction[- ]ready\b",
    r"(?i)\bguaranteed uptime\b",
    r"(?i)\bzero cost\b",
    r"(?i)\bsupports safe in-place downgrade\b",
    r"(?i)\bautomatic rollback is enabled\b",
    r"(?i)\bauto[- ]?updates are enabled\b",
    r"(?i)INFLUXDB3_EXTERNAL_BEARER_TOKEN\s*=\s*(?!\$\{\{secret|\")",
]:
    assert not re.search(forbidden, joined), forbidden

assert "ENTRYPOINT [\"/usr/local/bin/init.sh\"]" in (root / "Dockerfile").read_text()
assert "--http-bind 127.0.0.1:8182" in (root / "supervisor.sh").read_text()
assert "env -u INFLUXDB3_EXTERNAL_BEARER_TOKEN influxdb3 serve" in (
    root / "supervisor.sh"
).read_text()

print(f"claim and secret scan: PASS; files={len(text_files)}")
