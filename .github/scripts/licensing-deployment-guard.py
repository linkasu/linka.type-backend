"""Skip automatic publication only for the reviewed licensing snapshot.

The exemption matches only this snapshot directly above BASE; normal runtime
changes and explicit workflow_dispatch retain their publication behavior.
shortcut: this exemption is snapshot-bound; refresh BASE/hashes if rebased.
"""
import hashlib
import os
from pathlib import Path
import subprocess
import sys

BASE = '8b77a36b5b5f9d1ef9bf37e1d5e6c4daa4bcd6a8'
EXPECTED = {
    ".github/workflows/deploy.yml": "991c70ab11ea2fbce7e00a579d773c1a198575ccd119a9f84f263ddfe00c8827",
    "LICENSE": "3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986",
    "README.md": "79c37660bbbe840ccd8a4d43dff4bcc6dc099a79e16e7e0713fd80b9a6516f82"
}
SCRIPT = '.github/scripts/licensing-deployment-guard.py'


def licensing_only(parent, files):
    return parent == BASE and files == EXPECTED


def self_test():
    assert licensing_only(BASE, dict(EXPECTED))
    assert not licensing_only('another-parent', dict(EXPECTED))
    assert not licensing_only(BASE, {**EXPECTED, 'runtime.py': 'changed'})
    assert not licensing_only(BASE, {**EXPECTED, 'package.json': 'changed'})
    assert not licensing_only(BASE, {})
    print('licensing deployment guard: 5 checks passed')


def should_deploy():
    event = os.environ.get('GITHUB_EVENT_NAME')
    if event not in ('push', 'workflow_run', 'pull_request') or os.environ.get('GITHUB_REF_TYPE') == 'tag':
        return True
    parent = subprocess.check_output(['git', 'rev-parse', 'HEAD^'], text=True).strip()
    if parent != BASE:
        return True
    paths = subprocess.check_output(
        ['git', 'diff', '--name-only', '-z', BASE, 'HEAD']
    ).decode().split('\0')[:-1]
    if SCRIPT not in paths:
        return True
    files = {
        path: hashlib.sha256(Path(path).read_bytes()).hexdigest()
        for path in paths if path != SCRIPT
    }
    return not licensing_only(parent, files)


if __name__ == '__main__':
    if '--self-test' in sys.argv:
        self_test()
    else:
        deploy = should_deploy()
        print('deploy=' + str(deploy).lower())
        with open(os.environ['GITHUB_OUTPUT'], 'a') as output:
            output.write('deploy=' + str(deploy).lower() + '\n')
