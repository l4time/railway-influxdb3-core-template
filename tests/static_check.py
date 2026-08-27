#!/usr/bin/env python3
from pathlib import Path
import hashlib
import json
import re

root = Path(__file__).resolve().parents[1]

files = sorted(
    path.relative_to(root).as_posix()
    for path in root.rglob("*")
    if path.is_file()
    and path.relative_to(root).parts[0] not in {".git", "preflight"}
    and "__pycache__" not in path.parts
)

expected = [
    ".dockerignore",
    ".github/ISSUE_TEMPLATE/bug_report.yml",
    ".github/ISSUE_TEMPLATE/config.yml",
    ".github/labels.yml",
    ".github/workflows/ci.yml",
    "CHANGELOG.md",
    "CODE_OF_CONDUCT.md",
    "CONTRIBUTING.md",
    "Dockerfile",
    "LICENSE",
    "NOTICE.md",
    "PACKAGE_SHA256SUMS",
    "README.md",
    "SECURITY.md",
    "SOURCE_SHA256SUMS",
    "TRADEMARKS.md",
    "adapter.go",
    "assets/README.md",
    "assets/icon.svg",
    "docs/README.md",
    "docs/build-report.md",
    "docs/environment.md",
    "docs/marketplace-overview.md",
    "docs/operations.md",
    "docs/product-kit-completion.md",
    "docs/qa-checklist.md",
    "docs/railway-template-contract.md",
    "docs/source-provenance.md",
    "docs/support.md",
    "docs/template-inventory.md",
    "init.sh",
    "overview.md",
    "railway.json",
    "secureinit.go",
    "supervisor.sh",
    "testclient.go",
    "tests/claim_check.py",
    "tests/static_check.py",
]
assert files == expected, (set(files) - set(expected), set(expected) - set(files))

r5 = {
    "Dockerfile": "d9120ca139a6f2fe2d42e9bbdb17d18aecdd0a4bc73034fed40c4b99026eba2d",
    "adapter.go": "e05e78f2b1b46464bda6307f32fb48a357627d48639bab7f8a790ebc3b13bbc6",
    "init.sh": "300233a413cf0315b85376b887e39bc6d9f45bd6ef3c1f450ea2e8cd269c422b",
    "secureinit.go": "9ab2f095c5ea37ea5ad1f073c2ae4fb738b0837cd067eb6aff7825c6043791e6",
    "supervisor.sh": "0495034b39f3a776fd113dd325ed0816343ad27b3b211d0504a671d0c5a45c86",
    "testclient.go": "939d9ef5a7927240372b451bf24117bdb96f46dfecb0444b0ef7416bd7fd716b",
}
for name, digest in r5.items():
    assert hashlib.sha256((root / name).read_bytes()).hexdigest() == digest, name

config = json.loads((root / "railway.json").read_text())
assert config["build"] == {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile",
}
assert config["deploy"] == {
    "healthcheckPath": "/healthz",
    "healthcheckTimeout": 300,
    "numReplicas": 1,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10,
    "sleepApplication": False,
}

source = {}
for line in (root / "SOURCE_SHA256SUMS").read_text().splitlines():
    digest, name = line.split("  ", 1)
    assert re.fullmatch(r"[0-9a-f]{64}", digest), line
    assert name not in source
    source[name] = digest
assert source == r5

package = {}
for line in (root / "PACKAGE_SHA256SUMS").read_text().splitlines():
    digest, name = line.split("  ", 1)
    assert re.fullmatch(r"[0-9a-f]{64}", digest), line
    assert name not in package
    package[name] = digest
assert sorted(package) == [name for name in files if name != "PACKAGE_SHA256SUMS"]
for name, digest in package.items():
    assert hashlib.sha256((root / name).read_bytes()).hexdigest() == digest, name

framed = hashlib.sha256()
for name in files:
    framed.update(name.encode() + b"\0" + (root / name).read_bytes() + b"\0")
print(f"static contract: PASS; files={len(files)}; digest={framed.hexdigest()}")
